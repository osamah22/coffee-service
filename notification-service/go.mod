module github.com/osamah22/coffee-service/notification-service

go 1.26.2

require (
	github.com/osamah22/coffee-service/shared v0.0.0
	github.com/rabbitmq/amqp091-go v1.10.0
)

replace github.com/osamah22/coffee-service/shared => ../shared
