/**
 * Copyright (c) 2023.
 * Project: go-date-snowflake
 * File:       snowflake_test.go
 * Date:     2023/08/18 17:06:55
 * Author:  Colyll
 * QQ :      857859975
 *
 */
package snowflake

import (
	"github.com/redis/go-redis/v9"
	"testing"
)

func TestSnowflake_Id(t *testing.T) {
	client := redis.NewClient(&redis.Options{Addr: "localhost:6379", DB: 0})
	s := New(client)
	gotId, err := s.Id()
	if err != nil {
		t.Errorf("Id() error = %v", err)
		return
	}
	println(gotId)
}
