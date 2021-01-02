package rtc

import (
	"fmt"
	"strconv"

	"github.com/pion/ion-sfu/pkg/common"
)

type SSRCType uint32

func NewSSRCType(strSSRC string) SSRCType {
	num, _ := strconv.ParseInt(strSSRC, 10, 64)
	return SSRCType(num)
}

func (this SSRCType) ServiceID() common.ServiceIDType {
	return common.ServiceIDType(this)
}

func (this SSRCType) Decimal() string {
	return strconv.FormatInt(int64(this), 10)
}

func (this SSRCType) String() string {
	return fmt.Sprintf(
		"%d(0x%x 0x%x 0x%x 0x%x)",
		this.FullID(),
		this.MediaSrcID(),
		this.MediaSrcType(),
		this.MediaType(),
		this.MediaAttr(),
	)
}

func (this SSRCType) BaseSSRC() SSRCType {
	return SSRCType(this.Base())
}

func (this SSRCType) Base() uint32 {
	return uint32(this) & 0xFFFFFFF0
}

func (this SSRCType) FullID() uint32 {
	return uint32(this)
}

func (this SSRCType) Index() uint32 {
	return uint32(this) >> 7
}

func (this SSRCType) MediaSrcID() uint32 {
	return uint32(this) & 0xFFFFFF80
}

func (this SSRCType) MediaSrcType() uint32 {
	return uint32(this) & 0x00000040
}

func (this SSRCType) IsMediaSrcTypeServer() bool {
	v1 := uint32(this) & 0x00000040
	if v1 == 0x00000040 {
		return true
	}
	return false
}

func (this SSRCType) IsMediaSrcTypeClient() bool {
	v1 := uint32(this) & 0x00000040
	if v1 == 0x00000040 {
		return false
	}
	return true
}

func (this SSRCType) SetMediaSrcTypeServer() SSRCType {
	v1 := uint32(this) | 0x00000040
	return SSRCType(v1)
}

func (this SSRCType) SetMediaSrcTypeClient() SSRCType {
	v1 := uint32(this) & 0xFFFFFFBF
	return SSRCType(v1)
}

func (this SSRCType) MediaType() uint32 {
	return uint32(this) & 0x00000030
}

func (this SSRCType) SetMediaType(m uint32) SSRCType {
	v1 := uint32(this) & 0xFFFFFFCF
	v2 := (m & 0x03) << 4
	return SSRCType(v1 | v2)
}

func (this SSRCType) MediaAttr() uint32 {
	return uint32(this) & 0x0000000F
}

func NewSSRCTypeFromBuffer(buf []byte) SSRCType {
	if len(buf) != 4 {
		return SSRCType(0)
	}
	var tmp uint32 = 0
	tmp |= (uint32(buf[0]) << 24)
	tmp |= (uint32(buf[1]) << 16)
	tmp |= (uint32(buf[2]) << 8)
	tmp |= (uint32(buf[3]))

	return SSRCType(tmp)
}
