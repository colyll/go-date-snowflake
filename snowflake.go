// Package snowflake
/**
 * Copyright (c) 2023.
 * Project: go-date-snowflake
 * File:       snowflake.go
 * Date:     2023/08/18 12:46:09
 * Author:  Colyll
 * QQ :      857859975
 *
 */

/**
 * 由于雪花算法加上日期有26位长，所以修改缩小到22位，容量至9999-12-31，104.8W/ms。
 * |-------------------------------雪花算法(64bits)----------------------------------------|
 * |---补位(1bit)--|---当前时间戳毫秒(41bits)----|--机器ID(10bits)--|--序号(12bits)--|
 *
 *           |---------------------------------修改后(48bits)--------------------------------|
 *  日期 + |--补位(1bit)--|--每日第几毫秒(27bits)--|--机器ID(9bits)--|-序号(11bits)--|
 *
 * e.g 2023081838643360073728
 */

package snowflake

import (
	"bytes"
	"context"
	"github.com/redis/go-redis/v9"
	"gopkg.in/yaml.v3"
	"os"
	"strconv"
	"time"
)

type _config struct {
	Opt *Options `yaml:"snowflake"`
}

type Options struct {
	IdBits        int    `yaml:"idBits"` // 序号位宽
	MachineId     int64  `yaml:"machineId"`
	MachineIdBits int    `yaml:"machineIdBits"` // 机器ID位宽
	RegionId      int64  `yaml:"regionId"`      // 可选, 建议2^6 -1以内 依赖(RegionIdBits)
	RegionIdBits  int    `yaml:"regionIdBits"`  // 可选, 建议6以内, 生成字符串长度24位
	CachePrefix   string `yaml:"cachePrefix"`
}

type Snowflake struct {
	opt         *Options
	id          int64
	millisecond int64
	timestamp   int64
	dateString  string
	idKey       bytes.Buffer
	redis       *redis.Client
}

func New(redis *redis.Client) *Snowflake {
	opt := Options{}
	opt.init()

	return &Snowflake{opt: &opt, redis: redis}
}

func (opt *Options) init() {
	if opt.IdBits == 0 {
		opt.IdBits = 11
	}

	if opt.MachineIdBits == 0 {
		opt.MachineIdBits = 9
	}

	if opt.CachePrefix == "" {
		opt.CachePrefix = "colyll:snowflake:"
	}

	configBytes, err := os.ReadFile("./config.yaml")
	if err != nil {
		return
	}
	configStruct := _config{}
	err = yaml.Unmarshal(configBytes, &configStruct)
	if err != nil {
		return
	}
	opt.IdBits = configStruct.Opt.IdBits
	opt.MachineId = configStruct.Opt.MachineId
	opt.MachineIdBits = configStruct.Opt.MachineIdBits
	opt.RegionId = configStruct.Opt.RegionId
	opt.RegionIdBits = configStruct.Opt.RegionIdBits
	opt.CachePrefix = configStruct.Opt.CachePrefix

}

// Id
// 生成格式：2023081838643360073728
func (s *Snowflake) Id() (idString string, err error) {
	timestamp := time.Now().UnixMilli()

	// 系统时间被重置
	if timestamp < s.timestamp {
		panic("time is turned back !")
	}

	// 需要更新snowflake 时间
	if timestamp != s.timestamp {
		s.setTime()
	}

	// 更新snowflake id
	s.getIncrementId()

	// 位运算填充bit数据
	id := s.millisecond<<(s.opt.RegionIdBits+s.opt.MachineIdBits+s.opt.IdBits) |
		s.opt.RegionId<<(s.opt.MachineIdBits+s.opt.IdBits) |
		s.opt.MachineId<<s.opt.IdBits |
		s.id
	idString = s.dateString + strconv.FormatInt(id, 10)

	return idString, err
}

func (s *Snowflake) setTime() {
	t := time.Now().UnixMilli()
	s.timestamp = t
	s.millisecond = t % 86400000
	s.dateString = time.Now().Format("20060102")
	s.idKey.Reset()
	s.idKey.WriteString(s.opt.CachePrefix)
	s.idKey.WriteString(strconv.FormatInt(s.opt.RegionId, 10))
	s.idKey.WriteString(":")
	s.idKey.WriteString(strconv.FormatInt(s.opt.MachineId, 10))
	s.idKey.WriteString(":")
	s.idKey.WriteString(strconv.FormatInt(s.millisecond, 10))
	s.initId()
}

func (s *Snowflake) initId() {
	s.redis.SetNX(context.Background(), s.idKey.String(), -1, time.Second)
}

/*
*
更新id
*/
func (s *Snowflake) getIncrementId() {
	incr := s.redis.Incr(context.Background(), s.idKey.String())

	// 从redis获取的id, 超过序号容量最大值, 重新获取可用时间
	if incr.Val() > -1^(-1<<s.opt.IdBits) {
		s.waitNextTime()
		s.setTime()
	}
	s.id = incr.Val()
}

/*
*
获取可用毫秒值
*/
func (s *Snowflake) waitNextTime() {
	t := time.Now().UnixMilli()
	for t <= s.timestamp {
		t = time.Now().UnixMilli()
	}
}
