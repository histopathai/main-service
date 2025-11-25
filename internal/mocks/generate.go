//go:build generate
// +build generate

package mocks

//go:generate mockgen -destination=repository.go -package=mocks -source=../domain/port/repository.go
//go:generate mockgen -destination=service.go -package=mocks -source=../domain/port/service.go
//go:generate mockgen -destination=storage.go -package=mocks -source=../domain/port/storage.go
//go:generate mockgen -destination=event.go -package=mocks -source=../domain/port/event.go
//go:generate mockgen -destination=telemetry.go -package=mocks -source=../domain/port/telemetry.go
//go:generate mockgen -destination=worker.go -package=mocks -source=../domain/port/worker.go
