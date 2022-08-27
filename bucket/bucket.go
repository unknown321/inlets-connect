package bucket

import "time"

type Bucket struct {
	Quota         int64         `yaml:"quota"`
	Value         int64         `yaml:"value"`
	LastAccess    *time.Time    `yaml:"lastAccess,omitempty"`
	LimitDuration time.Duration `yaml:"limitDuration"`
}

type Buckets map[string]*Bucket

func (b *Bucket) ResetQuota() {
	b.Value = 0
}

func (b *Bucket) Init() *Bucket {
	if b.LastAccess == nil {
		t := time.Now()
		b.LastAccess = &t
	}

	return b
}
