package common

import (
	"fmt"
	"strconv"
	"strings"
)

//        0x00000040                                100 0000
//                                              mediaSrcType
//                                                  ^
//                                                  |
//                                                  |
//                   31     24   20   16   12    8  | 4    0
// center 0xFC000000 1111 1100 0000 0000 0000 0000 0000 0000  6
// domain 0x3FC0000         11 1111 1100 0000 0000 0000 0000  8
// type   0x3FC00                     11 1111 1100 0000 0000  8
// id     0x3FF                                 11 1111 1111  10   (id < 64 就不会有问题)
//
//
//#define SID_CENTER(SID) ((SID & 0xFC000000) >> 26)
//#define SID_DOMAIN(SID) ((SID & 0x3FC0000) >> 18)
//#define SID_TYPE(SID) ((SID & 0x3FC00) >> 10)
//#define SID_ID(SID) (SID & 0x3FF)
type ServiceIDType uint32

// serviceid like mas_13_1
func NewSServiceIDType(serviceID string, rtpRegion uint32) (ServiceIDType, error) {
	var s ServiceIDType
	items := strings.Split(serviceID, "_")
	if len(items) < 3 {
		return s, fmt.Errorf("len split id < 3")
	}
	center := 0
	//domain, err := strconv.Atoi(items[1])
	//domain, err := strconv.Atoi(idRegion)
	//if nil != err {
	//	return s, err
	//}
	stype, exist := _serviceNameMap[items[0]]
	if !exist {
		return s, fmt.Errorf("unkonw service name %v", items[0])
	}
	sid, err := strconv.Atoi(items[2])
	if nil != err {
		return s, err
	}

	domain := rtpRegion

	s = ServiceIDType((uint32(center) << 26) + (uint32(domain) << 18) + (uint32(stype << 10)) + uint32(sid))
	return s, nil
}

func (this ServiceIDType) String(idRegion string) string {
	return fmt.Sprintf(
		"%v_%v_%v",
		_serviceTypeMap[this.Type()],
		//this.Domain(),
		idRegion,
		this.ID(),
	)
}

func (this ServiceIDType) Center() uint32 {
	return (uint32(this) & 0xFC000000) >> 26
}

func (this ServiceIDType) Domain() uint32 {
	return (uint32(this) & 0x3FC0000) >> 18
}

func (this ServiceIDType) Type() uint32 {
	return (uint32(this) & 0x3FC00) >> 10
}

func (this ServiceIDType) ID() uint32 {
	return (uint32(this) & 0x3FF)
}

var _serviceNameMap map[string]uint32 = map[string]uint32{
	"UNKNOWN": 0,
	"mrs":     1,
	"mas":     2,
	"mcs":     3,
	"maps":    4,
	"rtcs":    5,
	"cc":      6,
	"mvps":    7,
	"rcs":     12,
	"mlg":     13,
}

var _serviceTypeMap map[uint32]string = map[uint32]string{
	0:  "UNKNOWN",
	1:  "mrs",
	2:  "mas",
	3:  "mcs",
	4:  "maps",
	5:  "rtcs",
	6:  "cc",
	7:  "mvps",
	12: "rcs",
	13: "mlg",
}

//typedef enum {
//  UNKNOWN = 0,
//  MRS = 1,
//  MAS = 2,
//  MCS = 3,
//  MAPS = 4,
//  RTCS = 5,
//  CALLROUTEGW = 6,
//  MVPS = 7,
//  SIPMGW = 8,
//  H5MGW = 9,
//  SIPSGW = 10,
//  H5SGW = 11,
//  RCS = 12,
//  MLG = 13,
//  WXMGW = 14,
//  MTS = 15
//} eServiceType;
