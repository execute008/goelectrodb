package electrodb

import (
	"time"
)

// TTL helper methods for setting time-to-live on items

// WithTTL sets a TTL (Time-To-Live) on a put operation
// The item will automatically be deleted by DynamoDB after the specified duration
func (p *PutOperation) WithTTL(duration time.Duration) *PutOperation {
	if p.entity.schema.TTL == nil {
		// No TTL configured in schema, silently ignore
		return p
	}

	ttlAttribute := p.entity.schema.TTL.Attribute
	ttlTimestamp := time.Now().Add(duration).Unix()
	p.item[ttlAttribute] = ttlTimestamp

	return p
}

// WithTTLTimestamp sets a TTL using an explicit Unix timestamp
func (p *PutOperation) WithTTLTimestamp(timestamp int64) *PutOperation {
	if p.entity.schema.TTL == nil {
		return p
	}

	ttlAttribute := p.entity.schema.TTL.Attribute
	p.item[ttlAttribute] = timestamp

	return p
}

// WithTTL sets a TTL (Time-To-Live) on an update operation
// The item will automatically be deleted by DynamoDB after the specified duration
func (u *UpdateOperation) WithTTL(duration time.Duration) *UpdateOperation {
	if u.entity.schema.TTL == nil {
		return u
	}

	ttlAttribute := u.entity.schema.TTL.Attribute
	ttlTimestamp := time.Now().Add(duration).Unix()
	u.setOps[ttlAttribute] = ttlTimestamp

	return u
}

// WithTTLTimestamp sets a TTL using an explicit Unix timestamp
func (u *UpdateOperation) WithTTLTimestamp(timestamp int64) *UpdateOperation {
	if u.entity.schema.TTL == nil {
		return u
	}

	ttlAttribute := u.entity.schema.TTL.Attribute
	u.setOps[ttlAttribute] = timestamp

	return u
}

// RemoveTTL removes the TTL from an item (prevents auto-deletion)
func (u *UpdateOperation) RemoveTTL() *UpdateOperation {
	if u.entity.schema.TTL == nil {
		return u
	}

	ttlAttribute := u.entity.schema.TTL.Attribute
	u.remOps = append(u.remOps, ttlAttribute)

	return u
}

// TTL utility functions

// TTLFromNow calculates a TTL timestamp from the current time plus duration
func TTLFromNow(duration time.Duration) int64 {
	return time.Now().Add(duration).Unix()
}

// TTLFromTime calculates a TTL timestamp from a specific time
func TTLFromTime(t time.Time) int64 {
	return t.Unix()
}

// IsTTLExpired checks if a TTL timestamp has expired
func IsTTLExpired(ttl int64) bool {
	return time.Now().Unix() > ttl
}

// TimeUntilTTL returns the duration until the TTL expires
func TimeUntilTTL(ttl int64) time.Duration {
	expirationTime := time.Unix(ttl, 0)
	return time.Until(expirationTime)
}
