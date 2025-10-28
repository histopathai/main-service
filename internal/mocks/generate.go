//go:build generate
// +build generate

package mocks

//go:generate mockgen -destination=repository.go -package=mocks -source=../domain/repository/interface.go
//go:generate mockgen -destination=storage.go -package=mocks -source=../domain/storage/interface.go
//go:generate mockgen -destination=publisher.go -package=mocks -source=../domain/events/interface.go Publisher,Subscriber,ImageEventPublisher
