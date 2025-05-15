// service/deleteservice/deleteservice.go
package deleteservice

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	selectservice "mymodule/service/proxyService/selectService"
	"mymodule/utils"

	"github.com/tidwall/sjson"
)

// Delete 从配置文件中删除指定 port 的 inbound。
// 返回 (true, "") 表示成功，否则 (false, 错误信息)。
func Delete(port int) (bool, string) {
	// 1. 读取加密文件
	rawEnc, err := os.ReadFile(utils.AES_File)
	if err != nil {
		return false, fmt.Sprintf("读取文件失败: %v", err)
	}
	// 2. 解密
	decrypted, err := utils.Decrypt(string(rawEnc))
	if err != nil {
		return false, fmt.Sprintf("解密失败: %v", err)
	}

	// 3. 拆 top-level JSON，取出 inbounds 数组
	var top map[string]json.RawMessage
	if err := json.Unmarshal(decrypted, &top); err != nil {
		return false, fmt.Sprintf("解析 JSON 失败: %v", err)
	}
	rawInb, ok := top["inbounds"]
	if !ok {
		return false, "inbounds 字段不存在"
	}
	var oldInbounds []json.RawMessage
	if err := json.Unmarshal(rawInb, &oldInbounds); err != nil {
		return false, fmt.Sprintf("inbounds 解析失败: %v", err)
	}

	// 4. 过滤掉要删除的 port
	filtered := make([]interface{}, 0, len(oldInbounds))
	for _, rawItem := range oldInbounds {
		var item struct {
			Port float64 `json:"port"`
		}
		if err := json.Unmarshal(rawItem, &item); err != nil {
			// 无法解析就保留
			filtered = append(filtered, rawItem)
			continue
		}
		if int(item.Port) == port {
			continue // 跳过
		}
		filtered = append(filtered, rawItem)
	}

	// 5. 用 sjson 写回 inbounds
	updated, err := sjson.SetBytes(decrypted, "inbounds", filtered)
	if err != nil {
		return false, fmt.Sprintf("sjson 替换失败: %v", err)
	}

	// 6. 格式化缩进（可选）
	var buf bytes.Buffer
	if err := json.Indent(&buf, updated, "", "  "); err != nil {
		// 如果缩进失败，就直接用 updated
		buf = *bytes.NewBuffer(updated)
	}

	// 7. 加密并写回
	enc, err := utils.Encrypt(buf.Bytes())
	if err != nil {
		return false, fmt.Sprintf("加密失败: %v", err)
	}
	if err := os.WriteFile(utils.AES_File, []byte(enc), 0o600); err != nil {
		return false, fmt.Sprintf("写文件失败: %v", err)
	}

	// 8. 删除后重置缓存
	selectservice.ResetCache()

	return true, ""
}
