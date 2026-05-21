package xray

import "fmt"

type Protocol string

const (
	ProtocolVLESS       Protocol = "vless"
	ProtocolTrojan      Protocol = "trojan"
	ProtocolShadowsocks Protocol = "shadowsocks"
)

// UserSpec 描述要在 Xray 入站里增删的用户。
// 不同协议用到的字段不同：VLESS 用 UUID + Flow，Trojan/Shadowsocks 用 Password。
type UserSpec struct {
	Email    string   // Xray 内唯一标识，约定走 EmailOf(userID)
	Level    uint32   // 0 即可，policy.levels 在此匹配
	Protocol Protocol // 默认 VLESS
	UUID     string   // VLESS / VMess
	Password string   // Trojan / Shadowsocks
	Flow     string   // VLESS：xtls-rprx-vision（Reality 推荐）
}

type TrafficStat struct {
	Uplink   int64
	Downlink int64
}

// EmailOf 把业务用户 ID 映射成 Xray 入站里的 email 标识。
// 用 ID 而不是 username，避免后续改名要同步 Xray。
func EmailOf(userID int64) string {
	return fmt.Sprintf("%d@routex", userID)
}
