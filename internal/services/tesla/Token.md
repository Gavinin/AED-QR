OAuth 登录流程
2021 年以后，Owner API 登录统一迁移到 Tesla SSO，并采用 PKCE。整体顺序如下：

生成 code_verifier、code_challenge 和随机 state。
在浏览器中打开授权页面，完成账号与 MFA 登录。
从重定向的地址栏取出授权码，换取 SSO 访问令牌。
使用 SSO 访问令牌兑换 Owner API 访问令牌。
定期使用刷新令牌续期。
1. 生成 PKCE 参数
```sh
code_verifier=$(openssl rand -base64 86 | tr -d '=+/\n' | cut -c1-86)
code_challenge=$(printf "%s" "$code_verifier" | openssl dgst -binary -sha256 | openssl base64 | tr '+/' '-_' | tr -d '=\n')
state=$(openssl rand -hex 12)
```

2. 构造授权请求
复制下方链接到浏览器，替换为自己的 code_challenge 与 state 后发起登录：

```
https://auth.tesla.com/oauth2/v3/authorize?client_id=ownerapi&code_challenge=${code_challenge}&code_challenge_method=S256&redirect_uri=https%3A%2F%2Fauth.tesla.com%2Fvoid%2Fcallback&response_type=code&scope=openid%20email%20offline_access&state=${state}
```

授权回调使用 https://auth.tesla.com/void/callback。这个地址本身会返回 404，但依然有存在意义：

Tesla 尚未向第三方开放 Owner API 的鉴权接口，只能复用官方 App 的 client_id，而该 client_id 绑定在这个回调域名上。
登录完成后虽然看到 404 页面，浏览器地址栏仍会携带 code，可以直接拿来换取访问令牌。
授权结束时，地址一般类似 https://auth.tesla.com/void/callback?code=...&state=...，把其中的 code 保存下来即可。如果账号归属中国区，需要把所有 auth.tesla.com 域名替换为 auth.tesla.cn，否则会遇到跨域或证书校验失败。

3. 用授权码换取 SSO 访问令牌
```
curl -X POST https://auth.tesla.com/oauth2/v3/token \
  -H 'Content-Type: application/json' \
  -d '{
    "grant_type": "authorization_code",
    "client_id": "ownerapi",
    "code": "<上一步的code>",
    "code_verifier": "'"$code_verifier"'",
    "redirect_uri": "https://auth.tesla.com/void/callback"
  }'
```

返回值包含 access_token（有效期约八小时）和 refresh_token（官方未公布期限，我的账号通常能用上数周）。

4. 交换 Owner API 访问令牌
```
curl -X POST https://owner-api.vn.teslamotors.com/oauth/token \
-H 'Content-Type: application/json' \
-d '{
"grant_type": "urn:ietf:params:oauth:grant-type:jwt-bearer",
"client_id": "ownerapi",
"client_secret": "<Owner API client_secret>",
"assertion": "<SSO access_token>"
}'
client_secret 可以在 Tim Dorr 的文档中找到，也可以自行抓包确认；若 Tesla 调整参数，只能再次想办法获取最新值。
```
5. 刷新令牌
```
   curl -X POST https://owner-api.vn.teslamotors.com/oauth/token \
   -H 'Content-Type: application/json' \
   -d '{
   "grant_type": "refresh_token",
   "client_id": "ownerapi",
   "client_secret": "<Owner API client_secret>",
   "refresh_token": "<Owner API refresh_token>"
   }'
```
我通常会在脚本即将执行或察觉登录态不稳时主动刷新，避免任务中途因为过期而失败。不少社区脚本也会在 access_token 即将过期时自动替换 新令牌，并据此重新检测区域，以免后续请求走错线路。

区域识别与刷新策略
实验下来，无论是 SSO 返回的 JWT 还是 Owner API 提供的刷新令牌，都携带了可识别区域的信息：iss 包含 .cn 的令牌基本属于中国区，全球线路的刷新令牌通常以 qts-、eu- 等前缀开头，中国区则会以 cn- 起头。我的记录显示，SSO access_token 大约 8 小时过期，refresh_token 则能撑到 45 天左右。

我在定时任务里解析这些标记来决定后续 API 请求的域名，并把刷新窗口控制在令牌到期前的 30～45 分钟，避免真正过期后再去拉起流程。

常用接口与数据采集
我的目标是记录行程，因此只保留与车辆状态相关的接口。

基础端点与请求头
REST API：北美与欧洲线路使用 https://owner-api.vn.teslamotors.com（老脚本里的 owner-api.teslamotors.com 目前仍然可用），中国线路对应 https://owner-api.vn.cloud.tesla.cn。
流式遥测：北美与欧洲线路使用 wss://streaming.vn.teslamotors.com/streaming/，中国线路对应 wss://streaming.vn.cloud.tesla.cn/streaming/。
认证域名：需要与账号所在区域匹配，分别是 https://auth.tesla.com 与 https://auth.tesla.cn。
每次调用都要带上 Authorization: Bearer <access_token>，顺手放一个稳定的 User-Agent（例如自定义的 MyTeslaScript/0.1.0）能帮忙排查故障。

REST API 轮询
常驻的几个接口如下：

GET /api/1/products：列出账号下的车辆与能源设备。/api/1/vehicles 现在已经无法使用了。
```json
{
"response": [
{
"id": 366525689971111,
"user_id": 1126490357715111,
"vehicle_id": 1126492832352111,
"vin": "LRWYGCFJ1SC927111",
"color": null,
"access_type": "OWNER",
"display_name": "小破电动车",
"option_codes": null,
"cached_data": "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",
"mobile_access_disabled": false,
"granular_access": {
"hide_private": false
},
"tokens": null,
"state": "offline",
"in_service": false,
"id_s": "366525689975782",
"calendar_enabled": true,
"api_version": 90,
"backseat_token": null,
"backseat_token_updated_at": null,
"ble_autopair_enrolled": false,
"device_type": "vehicle",// car!
"command_signing": "required",
"release_notes_supported": true
}
],
"count": 1
}
```

GET /api/1/vehicles/{id}/vehicle_data：获取车辆状态快照，我的定时任务完全依赖这一端点。
```json
// 车辆离线
{
"response": null,
"error": "vehicle unavailable: vehicle is offline or asleep",
"error_description": ""
},

// 车辆在线

{
  "response": {
    "id": 12345678901234567,
    "user_id": 123,
    "vehicle_id": 1234567890,
    "vin": "5YJSA11111111111",
    "display_name": "Nikola 2.0",
    "color": null,
    "access_type": "OWNER",
    "tokens": ["abcdef1234567890", "1234567890abcdef"],
    "state": "online",
    "in_service": false,
    "id_s": "12345678901234567",
    "calendar_enabled": true,
    "api_version": 13,
    "backseat_token": null,
    "backseat_token_updated_at": null,
    "drive_state": {
      "gps_as_of": 1607623884,
      "heading": 5,
      "latitude": 33.111111,
      "longitude": -88.111111,
      "native_latitude": 33.111111,
      "native_location_supported": 1,
      "native_longitude": -88.111111,
      "native_type": "wgs",
      "power": -9,
      "shift_state": null,
      "speed": null,
      "timestamp": 1607623897515
    },
    "climate_state": {
      "battery_heater": false,
      "battery_heater_no_power": false,
      "bioweapon_mode": false,
      "climate_keeper_mode": "off",
      "defrost_mode": 0,
      "driver_temp_setting": 21.1,
      "fan_status": 0,
      "inside_temp": 22.1,
      "is_auto_conditioning_on": false,
      "is_climate_on": false,
      "is_front_defroster_on": false,
      "is_preconditioning": false,
      "is_rear_defroster_on": false,
      "left_temp_direction": -66,
      "max_avail_temp": 28.0,
      "min_avail_temp": 15.0,
      "outside_temp": 18.0,
      "passenger_temp_setting": 21.1,
      "remote_heater_control_enabled": false,
      "right_temp_direction": -66,
      "seat_heater_left": 0,
      "seat_heater_right": 0,
      "side_mirror_heaters": false,
      "timestamp": 1607623897515,
      "wiper_blade_heater": false
    },
    "charge_state": {
      "battery_heater_on": false,
      "battery_level": 59,
      "battery_range": 149.92,
      "charge_current_request": 40,
      "charge_current_request_max": 40,
      "charge_enable_request": true,
      "charge_energy_added": 2.42,
      "charge_limit_soc": 90,
      "charge_limit_soc_max": 100,
      "charge_limit_soc_min": 50,
      "charge_limit_soc_std": 90,
      "charge_miles_added_ideal": 10.0,
      "charge_miles_added_rated": 8.0,
      "charge_port_cold_weather_mode": null,
      "charge_port_door_open": true,
      "charge_port_latch": "Engaged",
      "charge_rate": 28.0,
      "charge_to_max_range": false,
      "charger_actual_current": 40,
      "charger_phases": 1,
      "charger_pilot_current": 40,
      "charger_power": 9,
      "charger_voltage": 243,
      "charging_state": "Charging",
      "conn_charge_cable": "SAE",
      "est_battery_range": 132.98,
      "fast_charger_brand": "<invalid>",
      "fast_charger_present": false,
      "fast_charger_type": "<invalid>",
      "ideal_battery_range": 187.4,
      "managed_charging_active": false,
      "managed_charging_start_time": null,
      "managed_charging_user_canceled": false,
      "max_range_charge_counter": 0,
      "minutes_to_full_charge": 165,
      "not_enough_power_to_heat": false,
      "scheduled_charging_pending": false,
      "scheduled_charging_start_time": null,
      "time_to_full_charge": 2.75,
      "timestamp": 1607623897515,
      "trip_charging": false,
      "usable_battery_level": 59,
      "user_charge_enable_request": null
    },
    "gui_settings": {
      "gui_24_hour_time": false,
      "gui_charge_rate_units": "mi/hr",
      "gui_distance_units": "mi/hr",
      "gui_range_display": "Rated",
      "gui_temperature_units": "F",
      "show_range_units": true,
      "timestamp": 1607623897515
    },
    "vehicle_state": {
      "api_version": 13,
      "autopark_state_v2": "standby",
      "autopark_style": "standard",
      "calendar_supported": true,
      "car_version": "2020.48.10 f8900cddd03a",
      "center_display_state": 0, //0 ,7 ,8 ,9 means leave
      "df": 0,
      "dr": 0,
      "fd_window": 0,
      "fp_window": 0,
      "ft": 0,
      "homelink_device_count": 2,
      "homelink_nearby": true,
      "is_user_present": false,
      "last_autopark_error": "no_error",
      "locked": false, // is locked ? 
      "media_state": { "remote_control_enabled": true },
      "notifications_supported": true,
      "odometer": 57869.762487,
      "parsed_calendar_supported": true,
      "pf": 0,
      "pr": 0,
      "rd_window": 0,
      "remote_start": false,
      "remote_start_enabled": true,
      "remote_start_supported": true,
      "rp_window": 0,
      "rt": 0,
      "sentry_mode": false,
      "sentry_mode_available": true,
      "smart_summon_available": true,
      "software_update": {
        "download_perc": 0,
        "expected_duration_sec": 2700,
        "install_perc": 1,
        "status": "",
        "version": ""
      },
      "speed_limit_mode": {
        "active": false,
        "current_limit_mph": 85.0,
        "max_limit_mph": 90,
        "min_limit_mph": 50,
        "pin_code_set": false
      },
      "summon_standby_mode_enabled": false,
      "sun_roof_percent_open": 0,
      "sun_roof_state": "closed",
      "timestamp": 1607623897515,
      "valet_mode": false,
      "valet_pin_needed": true,
      "vehicle_name": null
    },
    "vehicle_config": {
      "can_accept_navigation_requests": true,
      "can_actuate_trunks": true,
      "car_special_type": "base",
      "car_type": "models2",
      "charge_port_type": "US",
      "default_charge_to_max": false,
      "ece_restrictions": false,
      "eu_vehicle": false,
      "exterior_color": "White",
      "has_air_suspension": true,
      "has_ludicrous_mode": false,
      "motorized_charge_port": true,
      "plg": true,
      "rear_seat_heaters": 0,
      "rear_seat_type": 0,
      "rhd": false,
      "roof_color": "None",
      "seat_type": 2,
      "spoiler_type": "None",
      "sun_roof_installed": 2,
      "third_row_seats": "None",
      "timestamp": 1607623897515,
      "trim_badging": "p90d",
      "use_range_badging": false,
      "wheel_type": "AeroTurbine19"
    }
  }
}

```
POST /api/1/vehicles/{id}/wake_up：调试时手动唤醒，正式运行时尽量避免，减少能耗。
```json
//API 会立即返回响应，但车辆实际上线并准备好接收其他指令可能需要几秒钟时间。一种解决方法是循环调用此端点，直到返回状态为“在线”为止，并设置超时时间。在某些情况下，唤醒过程可能较慢，因此建议使用至少 30 秒的超时时间。

{
"response": {
"id": 12345678901234567,
"user_id": 12345,
"vehicle_id": 1234567890,
"vin": "5YJSA11111111111",
"display_name": "Nikola 2.0",
"color": null,
"tokens": ["abcdef1234567890", "1234567890abcdef"],
"state": "online",
"in_service": false,
"id_s": "12345678901234567",
"calendar_enabled": true,
"api_version": 7,
"backseat_token": null,
"backseat_token_updated_at": null
}
}
```

## Doors

### POST `/api/1/vehicles/{id}/command/door_unlock`

Unlocks the doors to the car. Extends the handles on the S.

#### Response

```json
{
  "reason": "",
  "result": true
}
```

### POST `/api/1/vehicles/{id}/command/door_lock`

Locks the doors to the car. Retracts the handles on the S, if they are extended.

#### Response

```json
{
  "reason": "",
  "result": true
}
```

## Frunk/Trunk

### POST `/api/1/vehicles/{id}/command/actuate_trunk`

Opens either the front or rear trunk. On the Model S and X, it will also close the rear trunk.

#### Parameters

| Parameter   | Example | Description                                                         |
| :---------- | :------ | :------------------------------------------------------------------ |
| which_trunk | rear    | Which trunk to open/close. `rear` and `front` are the only options. |

#### Response

```json
{
  "reason": "",
  "result": true
}
```

🌐 线路差异：若账号在中国区，域名需要改为 https://owner-api.vn.cloud.tesla.cn。我自用的脚本会根据账号所在区域切换域名，同时把 OAuth 流程中的 auth.tesla.com 也替换成 auth.tesla.cn。

接口返回的每辆车会同时包含 id（REST 调用使用）与 vehicle_id（流式遥测使用）。社区经验普遍提醒不要混淆这两个字段。调用 GET /api/1/vehicles/{id}/vehicle_data 时，也可以通过 endpoints=drive_state;charge_state 等参数裁剪返回内容，减少带宽与解析压力。

定时抓取车辆状态
我会周期性调用 GET /api/1/vehicles/{id}/vehicle_data，把整车状态的快照写进数据库。返回值涵盖位置、电量、空调、轮胎压力、锁车状态等字段，足够用于行程回放。写入后，我会观察档位：如果车辆持续处于 P 挡超过五分钟，就认定行程已经结束，随后生成摘要推送到 Telegram，包括起止时间、地理位置、电量变化和平均速度。

REST 接口的数值大多遵循北美的英制单位：drive_state.speed 以 mph 表示车速，odometer 与相关里程字段使用 mile。为了在日常记录里保留公里与公里/小时等公制指标，我在入库前做一次换算，例如把 mph 乘以 1.609344 转换为 km/h，再根据需要保留两位小数。

服务端会以 msg_type=data:update 推送一行逗号分隔的字符串，频率通常是一秒一次。字段顺序固定为时间戳、车速、里程、电量、海拔、方向、纬度、经度、功率、档位、续航、估算续航和航向角；常见的参考资料给出了常用单位，例如车速以 mph、里程以 mile、电量为百分比，功率以 kW 计算。我会在解析阶段顺便把 mph 与 mile 转换成 km/h 与 km，保证统计口径与其他驾驶记录保持一致。车辆离线或蜂窝网络断开时，连接不会自动关闭，只是长时间没有新消息。如果客户端没有超时逻辑，就会一直等待。TeslaMate 的做法是 30 秒内没有更新就主动断开 WebSocket，改为每 30 秒调用 JSON API；该接口会明确返回 state（例如 offline），便于判断当前状态。我在自用脚本里也沿用这一策略：超过 30 秒无数据就切回 REST 轮询，待车辆重新上线后再恢复流式订阅。

错误与重试
REST 遇到问题时返回的信息不尽相同。常见的 HTTP 状态码包括 401（令牌过期需刷新）、403（账号权限不足）、429（触发限流，按指数退避重试）以及 451（请求落在错误区域）。流式接口偶尔会返回 data:error，其中的 vehicle_disconnected 或 Can't validate token 等提示能帮助定位问题。

为了降低噪声，我把失败的请求做指数退避，最长等待 30 分钟；WebSocket 则增加心跳检测，避免假在线状态。不少脚本也会按照类似策略自动重连。