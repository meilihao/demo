package main

import (
	"context"
	//slog "log"
	"os"
	"os/signal"

	"github.com/meilihao/golib/v2"
	"github.com/meilihao/golib/v2/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"moul.io/zapgorm2"
)

type Product struct {
	gorm.Model
	Code   string
	Price  uint
	Author Author `gorm:"type:JSON;serializer:json"`
}

type Author struct {
	Name  string
	Email string
}

// TODO: logger 与 zap level不一致
func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	log.SetDefaultLevel("debug")
	shutdownFn, err := golib.InitOTEL(":4317", "t", log.Glog, log.GSlog)
	if err != nil {
		panic(err)
	}
	defer shutdownFn(ctx)

	tracer := otel.Tracer("orm")

	ctx, span := tracer.Start(ctx, "test")
	defer span.End()

	// newLogger := logger.New(
	// 	slog.New(os.Stdout, "\r\n", slog.LstdFlags), // io writer
	// 	logger.Config{
	// 		SlowThreshold: time.Second, // 慢 SQL 阈值
	// 		LogLevel:      logger.Info, // Log level
	// 		Colorful:      false,       // 禁用彩色打印
	// 	},
	// )

	newLogger := zapgorm2.New(log.Glog)
	newLogger.LogLevel = logger.Silent
	newLogger.Context = func(ctx context.Context) []zapcore.Field {
		// github.com/uptrace/opentelemetry-go-extra/otelzap@v0.2.3/otelzap.go
		fields := make([]zapcore.Field, 0, 5)

		span := trace.SpanFromContext(ctx)
		if !span.IsRecording() {
			return fields
		}

		spanCtx := span.SpanContext()
		fields = append(fields, zap.String("trace_id", spanCtx.TraceID().String()),
			zap.String("span_id", spanCtx.SpanID().String()),
		)

		return fields
	}
	newLogger.SetAsDefault()

	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
			NoLowerCase:   true,
		},
		Logger: newLogger,
	})
	if err != nil {
		panic("failed to connect database")
	}

	if err := db.Use(otelgorm.NewPlugin()); err != nil { // 生成新span并发送到了otelcol-contrib
		panic(err)
	}

	zlog := otelzap.New(zap.NewExample(),
		otelzap.WithTraceIDField(true),
		otelzap.WithMinLevel(zap.DebugLevel),
	)

	// db.AutoMigrate(&Product{})

	// p1 := &Product{Code: "D1", Price: 100, Author: Author{
	// 	Name: "chen",
	// }}
	// if err = db.Create(p1).Error; err != nil {
	// 	log.Println(err)
	// }

	// p2 := &Product{Code: "D2", Price: 100}
	// if err = db.Create(p2).Error; err != nil {
	// 	log.Println(err)
	// }
	ls := []*Product{}
	if err = db.Debug().WithContext(ctx).Find(&ls).Error; err != nil {
		panic(err)
	}
	zlog.Ctx(ctx).Info("hello from zap", zap.Any("ls", ls))
}
