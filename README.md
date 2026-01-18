# Home Gateway - 瀹跺涵鏅鸿兘鎺у埗API缃戝叧

涓€涓繍琛屽湪 ARM64 璁惧锛堝 Rock 5B锛変笂鐨勬櫤鑳藉灞呮帶鍒剁綉鍏筹紝閫氳繃 LLM 杩涜鑷劧璇█鎰忓浘璇嗗埆銆?

## 鍔熻兘鐗规€?

- 馃 **LLM 鏅鸿兘璇嗗埆**: 浣跨敤澶ц瑷€妯″瀷璇嗗埆鐢ㄦ埛鎰忓浘锛屽尮閰嶅悎閫傜殑澶勭悊鍣?
- 馃摫 **澶氭笭閬撴敮鎸?*: 鏀寔 HTTP API銆乀elegram銆佷紒涓氬井淇＄瓑澶氱杈撳叆娓犻亾
- 馃敡 **鍙墿灞曞鐞嗗櫒**: 閫氳繃閰嶇疆鏂囦欢瀹氫箟澶勭悊鍣紝鏃犻渶淇敼浠ｇ爜
- 馃摠 **Kafka 娑堟伅闃熷垪**: 涓庡悗绔鐞嗗櫒閫氳繃 Kafka 寮傛閫氫俊
- 馃攧 **閰嶇疆鐑噸杞?*: 鏀寔杩愯鏃堕噸鏂板姞杞介厤缃?
- 馃摝 **鑷姩鏇存柊**: 鍐呯疆鑷洿鏂板姛鑳斤紝涓€閿崌绾?

## 蹇€熷紑濮?

### 1. 涓嬭浇

浠?[Releases](https://github.com/yoyo3287258/home-gateway/releases) 椤甸潰涓嬭浇瀵瑰簲骞冲彴鐨勪簩杩涘埗鏂囦欢锛?

```bash
# Rock 5B / ARM64 Linux
wget https://github.com/yoyo3287258/home-gateway/releases/latest/download/home-gateway-linux-arm64
chmod +x home-gateway-linux-arm64
```

### 2. 閰嶇疆

涓嬭浇閰嶇疆鏂囦欢绀轰緥锛?

```bash
wget https://github.com/yoyo3287258/home-gateway/releases/latest/download/configs-example.tar.gz
tar -xzf configs-example.tar.gz
```

缂栬緫 `configs/config.yaml`锛岃缃綘鐨?LLM API 瀵嗛挜锛?

```yaml
llm:
  base_url: "https://api.aihubmix.com/v1"  # 鎴栧叾浠?OpenAI 鍏煎鐨?API
  api_key: "your-api-key"
  model: "gpt-4o-mini"
```

### 3. 杩愯

```bash
# 璁剧疆鐜鍙橀噺锛堟帹鑽愶級
export LLM_API_KEY="your-api-key"
export LLM_BASE_URL="https://api.aihubmix.com/v1"

# 杩愯
./home-gateway-linux-arm64
```

### 4. 娴嬭瘯

```bash
# 鍋ュ悍妫€鏌?
curl http://localhost:8080/api/v1/health

# 鍙戦€佹寚浠?
curl -X POST http://localhost:8080/api/v1/command \
  -H "Content-Type: application/json" \
  -d '{"text": "鎵撳紑瀹㈠巺鐨勭伅"}'
```

## 鍛戒护琛屽弬鏁?

```bash
./home-gateway [閫夐」]

閫夐」:
  -c, -config string       涓婚厤缃枃浠惰矾寰?(榛樿 "configs/config.yaml")
  -p, -processors string   澶勭悊鍣ㄩ厤缃枃浠惰矾寰?(榛樿 "configs/processors.yaml")
  -v, -version            鏄剧ず鐗堟湰淇℃伅
  -U, -update             妫€鏌ュ苟鏇存柊鍒版渶鏂扮増鏈?
```

## API 鎺ュ彛

| 鏂规硶 | 璺緞 | 璇存槑 |
|------|------|------|
| GET | `/api/v1/health` | 鍋ュ悍妫€鏌?|
| GET | `/api/v1/processors` | 鑾峰彇澶勭悊鍣ㄥ垪琛?|
| POST | `/api/v1/command` | 鍙戦€佹帶鍒舵寚浠?|
| POST | `/api/v1/config/reload` | 閲嶆柊鍔犺浇閰嶇疆 |
| POST | `/api/v1/webhook/telegram` | Telegram Webhook |

### 鍙戦€佹寚浠ょず渚?

**璇锋眰:**
```json
{
  "text": "鎶婂鍘呯伅璋冩殫涓€鐐?,
  "channel": "http"
}
```

**鍝嶅簲:**
```json
{
  "success": true,
  "message": "瀹㈠巺鐏凡璋冩殫鑷?0%",
  "trace_id": "abc123",
  "processor_id": "light_control",
  "data": {
    "current_brightness": 50
  }
}
```

## 閰嶇疆璇存槑

### 澶勭悊鍣ㄩ厤缃?(processors.yaml)

```yaml
processors:
  - id: "light_control"
    name: "鐏厜鎺у埗"
    description: "鎺у埗瀹朵腑鍚勬埧闂寸殑鐏厜"
    keywords: ["鐏?, "寮€鐏?, "鍏崇伅"]
    parameters:
      - name: "room"
        type: "string"
        required: true
        description: "鎴块棿鍚嶇О"
      - name: "action"
        type: "enum"
        required: true
        values: ["on", "off", "dim"]
        description: "鎿嶄綔绫诲瀷"
```

## Kafka 娑堟伅鏍煎紡

### 璇锋眰娑堟伅 (Gateway 鈫?Processor)

```json
{
  "trace_id": "uuid",
  "timestamp": "2026-01-18T13:30:00+08:00",
  "processor_id": "light_control",
  "parameters": {
    "room": "瀹㈠巺",
    "action": "on"
  },
  "original_text": "鎵撳紑瀹㈠巺鐨勭伅",
  "channel": "telegram"
}
```

### 鍝嶅簲娑堟伅 (Processor 鈫?Gateway)

```json
{
  "trace_id": "uuid",
  "timestamp": "2026-01-18T13:30:01+08:00",
  "processor_id": "light_control",
  "success": true,
  "message": "瀹㈠巺鐏凡鎵撳紑"
}
```

## 寮€鍙?

### 鏈湴鏋勫缓

```bash
# 瀹夎渚濊禆
go mod download

# 杩愯
go run ./cmd/gateway

# 鏋勫缓
go build -o home-gateway ./cmd/gateway

# 浜ゅ弶缂栬瘧 (ARM64)
GOOS=linux GOARCH=arm64 go build -o home-gateway-linux-arm64 ./cmd/gateway
```

### 杩愯娴嬭瘯

```bash
go test -v ./...
```

## License

MIT License
