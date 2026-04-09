package generator

import (
	"fmt"
	"sync"
	"time"
)

// SnowflakeGenerator генерирует уникальные 64-битные ID
type SnowflakeGenerator struct {
	mu       sync.Mutex
	lastTime int64 // последний использованный timestamp
	sequence int64 // счётчик в пределах одной миллисекунды
	nodeID   int64 // ID узла (0-1023)
	epoch    int64 // кастомная эпоха (в миллисекундах)
}

// Константы для битовых сдвигов
// Структура ID: [41 бит timestamp][10 бит nodeID][12 бит sequence]
const (
	defaultEpoch = int64(1767225600000)                // 2026-01-01 00:00:00 UTC
	nodeBits     = uint(10)                            // 10 бит для ID узла (1024 уникальных узла)
	seqBits      = uint(12)                            // 12 бит для счётчика (4096 значений на миллисекунду)
	nodeMax      = int64(-1) ^ (int64(-1) << nodeBits) // 1023
	seqMax       = int64(-1) ^ (int64(-1) << seqBits)  // 4095
	timeShift    = nodeBits + seqBits                  // 22 бита сдвига для времени
	nodeShift    = seqBits                             // 12 бит сдвига для nodeID
)

// NewSnowflakeGenerator создаёт новый генератор
func NewSnowflakeGenerator(nodeID int64) (*SnowflakeGenerator, error) {
	if nodeID < 0 || nodeID > nodeMax {
		return nil, fmt.Errorf("nodeID must be between 0 and %d", nodeMax)
	}

	return &SnowflakeGenerator{
		nodeID: nodeID,
		epoch:  defaultEpoch,
	}, nil
}

// Генерирует уникальный 64-битный ID
func (s *SnowflakeGenerator) Generate() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UnixMilli()

	// Если время совпадает с предыдущим — увеличиваем счётчик
	if now == s.lastTime {
		s.sequence = (s.sequence + 1) & seqMax
		if s.sequence == 0 {
			// Переполнение sequence
			// Ждём следующую миллисекунду
			for now <= s.lastTime {
				now = time.Now().UnixMilli()
			}
		}
	} else {
		s.sequence = 0
	}

	// Защита от отката часов
	if now < s.lastTime {
		panic("clock moved backwards")
	}

	s.lastTime = now

	// Сборка ID: timestamp | nodeID | sequence
	return ((now - s.epoch) << timeShift) | (s.nodeID << nodeShift) | s.sequence
}

// разбирает ID на составные части
func (s *SnowflakeGenerator) Decompose(id int64) (timestamp int64, nodeID int64, sequence int64) {
	timestamp = (id >> timeShift) + s.epoch
	nodeID = (id >> nodeShift) & nodeMax
	sequence = id & seqMax
	return
}
