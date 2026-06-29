# Проверка и профилирование HW10

Все команды выполняются из директории домашнего задания:

```bash
cd /Users/artemnovikov/Developer/home_work/hw10_program_optimization
```

## Подготовка

Загрузить зависимости:

```bash
go mod download
```

Установить `benchstat` для сравнения результатов:

```bash
go install golang.org/x/perf/cmd/benchstat@latest
```

## Форматирование и статический анализ

```bash
gofmt -w stats.go stats_test.go stats_benchmark_test.go
go vet ./...
golangci-lint run
```

Если локальный `golangci-lint` несовместим с конфигурацией репозитория:

```bash
go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.63.4 run
```

## Обычные тесты

Файлы с `//go:build !bench` проверяются без тега `bench`:

```bash
go test -v -count=1 ./...
```

Проверка гонок данных:

```bash
go test -race -count=1 ./...
```

Покрытие:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
go tool cover -html=coverage.out
```

## Проверка лимитов задания

Запускает `TestGetDomainStat_Time_And_Memory` из файла с
`//go:build bench` и проверяет результат, время и память:

```bash
go test -v -count=1 -timeout=30s -tags bench .
```

Требования теста:

```text
time   < 300 ms
memory < 30 MiB
result == expectedBizStat
```

## Benchmark в памяти

ZIP распаковывается один раз до запуска таймера. Измеряется только обработка
данных функцией `GetDomainStat`.

Новая оптимизированная реализация:

```bash
go test -tags bench -run '^$' \
  -bench '^BenchmarkGetDomainStat$' \
  -benchmem -benchtime=2s -count=10 . | tee new.txt
```

Предыдущая однопоточная `fastjson`-реализация:

```bash
HW10_BENCH_IMPL=old go test -tags bench -run '^$' \
  -bench '^BenchmarkGetDomainStat$' \
  -benchmem -benchtime=2s -count=10 . | tee old-fastjson.txt
```

Исходная неоптимизированная реализация задания:

```bash
HW10_BENCH_IMPL=original go test -tags bench -run '^$' \
  -bench '^BenchmarkGetDomainStat$' \
  -benchmem -benchtime=2s -count=10 . | tee original.txt
```

## Benchmark с ZIP-декомпрессией

Новая реализация:

```bash
go test -tags bench -run '^$' \
  -bench '^BenchmarkGetDomainStatZIP$' \
  -benchmem -benchtime=3x -count=10 . | tee new-zip.txt
```

Исходная реализация:

```bash
HW10_BENCH_IMPL=original go test -tags bench -run '^$' \
  -bench '^BenchmarkGetDomainStatZIP$' \
  -benchmem -benchtime=3x -count=10 . | tee original-zip.txt
```

## Сравнение через benchstat

Исходная и новая реализации:

```bash
$(go env GOPATH)/bin/benchstat original.txt new.txt
```

Предыдущая `fastjson` и новая реализации:

```bash
$(go env GOPATH)/bin/benchstat old-fastjson.txt new.txt
```

ZIP-сценарий:

```bash
$(go env GOPATH)/bin/benchstat original-zip.txt new-zip.txt
```

CSV без рамок для дальнейшего форматирования:

```bash
$(go env GOPATH)/bin/benchstat -format csv original.txt new.txt \
  2>/dev/null | column -s, -t
```

Для доверительного интервала нужен `-count` не меньше 6. В примерах
используется `-count=10`.

## CPU и memory profile

Собрать отдельный test binary:

```bash
go test -c -tags bench -o /tmp/hw10.test .
```

Снять профили ускоренной реализации:

```bash
/tmp/hw10.test \
  -test.run '^$' \
  -test.bench '^BenchmarkGetDomainStat$' \
  -test.benchtime=3s \
  -test.cpuprofile=/tmp/hw10-cpu.pprof \
  -test.memprofile=/tmp/hw10-mem.pprof
```

Посмотреть основные потребители CPU и памяти:

```bash
go tool pprof -top /tmp/hw10.test /tmp/hw10-cpu.pprof
go tool pprof -top -alloc_space /tmp/hw10.test /tmp/hw10-mem.pprof
```

Открыть интерактивный web-интерфейс CPU profile:

```bash
go tool pprof -http=:8080 /tmp/hw10.test /tmp/hw10-cpu.pprof
```