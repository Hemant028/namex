package analytics

import (
	"context"
	"log"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
)

type RequestLog struct {
	Timestamp   time.Time
	DomainID    int
	ClientIP    string
	UserAgent   string
	Method      string
	Path        string
	Status      int
	Duration    int64 // Microseconds
	ActionTaken string
}

type Repository interface {
	Log(req *RequestLog)
	Close()
}

type clickHouseRepository struct {
	conn       clickhouse.Conn
	logChan    chan *RequestLog
	batchSize  int
	flushTimer time.Duration
	done       chan struct{}
}

func NewRepository(conn clickhouse.Conn) Repository {
	repo := &clickHouseRepository{
		conn:       conn,
		logChan:    make(chan *RequestLog, 10000),
		batchSize:  1000,
		flushTimer: 5 * time.Second,
		done:       make(chan struct{}),
	}
	go repo.worker()
	return repo
}

func (r *clickHouseRepository) Log(req *RequestLog) {
	select {
	case r.logChan <- req:
	default:
		// Drop log if channel is full to avoid blocking
		log.Println("Analytics channel full, dropping log")
	}
}

func (r *clickHouseRepository) Close() {
	close(r.logChan)
	<-r.done
}

func (r *clickHouseRepository) worker() {
	defer close(r.done)

	batch := make([]*RequestLog, 0, r.batchSize)
	ticker := time.NewTicker(r.flushTimer)
	defer ticker.Stop()

	flush := func() {
		if len(batch) == 0 {
			return
		}
		if err := r.writeBatch(batch); err != nil {
			log.Printf("Failed to write analytics batch: %v", err)
		}
		batch = make([]*RequestLog, 0, r.batchSize)
	}

	for {
		select {
		case req, ok := <-r.logChan:
			if !ok {
				flush()
				return
			}
			batch = append(batch, req)
			if len(batch) >= r.batchSize {
				flush()
			}
		case <-ticker.C:
			flush()
		}
	}
}

func (r *clickHouseRepository) writeBatch(logs []*RequestLog) error {
	ctx := context.Background()
	batch, err := r.conn.PrepareBatch(ctx, "INSERT INTO requests")
	if err != nil {
		return err
	}

	for _, l := range logs {
		err := batch.Append(
			l.Timestamp,
			int32(l.DomainID),
			l.ClientIP,
			l.UserAgent,
			l.Method,
			l.Path,
			int32(l.Status),
			l.Duration,
			l.ActionTaken,
		)
		if err != nil {
			return err
		}
	}

	return batch.Send()
}
