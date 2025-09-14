package kafka

type KafkaConfig struct {
	Brokers  []string
	Topic    string
	Group_id string
}

type Msg struct {
	OrderId int64
	Driver  Driver
}

type Driver struct {
	ID            int64  `json:"id" db:"id"`
	Name          string `json:"name" db:"name"`
	Phone         string `json:"phone" db:"phone"`
	LicenseNumber string `json:"license_number" db:"license_number"`
	Car           string `json:"car" db:"car"`
}
