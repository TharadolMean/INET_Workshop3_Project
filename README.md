# Employee Profile Management API

REST API สำหรับจัดการโปรไฟล์พนักงาน ส่งเป็น Final Project การอบรม Golang ที่ INET ใช้ Go, Fiber v2, GORM และ SQLite ไม่มี frontend

## Technology stack และ features

- Go, Fiber v2, GORM, SQLite (pure Go driver `modernc.org/sqlite`), UUID
- CRUD ด้วย public `employee_id` แบบ UUID; `gorm.Model.ID` ใช้ภายในฐานข้อมูล
- HTTP Basic Authentication สำหรับ POST, PUT, DELETE; GET ทุก endpoint อ่านได้โดยไม่ต้อง login
- Validation ข้อมูลก่อน create/update, ค้นหาแบบ partial match, จัดกลุ่มอายุตามโจทย์
- AutoMigrate และ seed ข้อมูลสมมติ 25 คนเมื่อฐานข้อมูลว่าง
- JSON response, Go integration tests และ Postman Collection

## Project structure

```text
.
├── main.go                    # เริ่มฐานข้อมูล, seed, server
├── config/database.go         # เชื่อมต่อ SQLite และ AutoMigrate
├── models/employee.go         # Employee และ public response
├── handlers/employee_handler.go
├── routes/routes.go
├── middleware/basic_auth.go
├── utils/validation.go
├── seed/employee_seed.go
├── main_test.go
├── docs/postman_collection.json
├── .env.example
├── .gitignore
├── go.mod
└── go.sum
```

## Installation requirements และขั้นตอน

1. ติดตั้ง Go **1.26 หรือใหม่กว่า** บน Windows และตรวจสอบ `go version` (โปรเจกต์นี้พัฒนาด้วย Go 1.27.2)
2. Clone repository แล้วเปิด PowerShell ในโฟลเดอร์โปรเจกต์
3. ตั้งค่า `$env:CGO_ENABLED = '0'` ใน PowerShell แล้วรัน `go mod download`
4. รัน `go run .`
5. เรียก API ที่ `http://localhost:3000/api/v1`

โปรเจกต์ใช้ `modernc.org/sqlite` ซึ่งเป็น pure Go และตั้ง `CGO_ENABLED=0` เพื่อให้ build ได้โดยไม่ต้องติดตั้ง GCC (GORM SQLite dialector มี dependency ที่รองรับ CGO แต่โปรเจกต์นี้ไม่ใช้ driver นั้นตอนรัน). หาก Go ติดตั้งที่ `C:\Program Files\Go` แต่คำสั่ง `go` ยังใช้ไม่ได้ ให้เพิ่ม `C:\Program Files\Go\bin` ใน PATH แล้วเปิด PowerShell ใหม่

## Database setup และ environment configuration

เมื่อเริ่มโปรแกรม GORM จะสร้างตารางด้วย AutoMigrate และ seed 25 คนเฉพาะเมื่อไม่มีพนักงานเลย การ restart ไม่เพิ่มข้อมูลซ้ำ ไฟล์ฐานข้อมูลเริ่มต้นคือ `employee.db` ในโฟลเดอร์ที่รันโปรแกรม และถูก ignore โดย Git

ตัวแปร environment ทั้งหมดเป็น optional (`.env.example` เป็นตัวอย่าง; โปรแกรมอ่าน environment โดยตรง ไม่โหลด `.env` อัตโนมัติ):

| Variable | Default | Purpose |
|---|---|---|
| `DB_PATH` | `employee.db` | ตำแหน่งไฟล์ SQLite |
| `BASIC_AUTH_USER` | `testgo` | Username สำหรับ endpoint ที่แก้ไขข้อมูล |
| `BASIC_AUTH_PASSWORD` | `23012023` | Password สำหรับ endpoint ที่แก้ไขข้อมูล |

ตัวอย่าง PowerShell:

```powershell
$env:DB_PATH = 'employee.db'
$env:BASIC_AUTH_USER = 'testgo'
$env:BASIC_AUTH_PASSWORD = '23012023'
$env:CGO_ENABLED = '0'
go run .
```

## API endpoints

| Method | Endpoint | Auth | Description |
|---|---|---|---|
| POST | `/api/v1/users` | Basic | สร้างพนักงาน |
| GET | `/api/v1/users` | ไม่ต้องใช้ | รายชื่อทั้งหมด |
| GET | `/api/v1/users/:id` | ไม่ต้องใช้ | ดูด้วย UUID |
| PUT | `/api/v1/users/:id` | Basic | แก้ไขข้อมูลทั้ง 5 ช่องที่ client กรอก |
| DELETE | `/api/v1/users/:id` | Basic | ลบพนักงาน |
| GET | `/api/v1/users/generations` | ไม่ต้องใช้ | จัดกลุ่มอายุ |
| GET | `/api/v1/users/search/:keyword` | ไม่ต้องใช้ | ค้นหาใน UUID, name, lastname |

Basic Auth ใช้ `testgo` / `23012023` ตามโจทย์ (หรือค่าจาก environment) หากไม่ส่งหรือส่งผิดจะได้ `401` พร้อม `WWW-Authenticate`. GET ไม่ต้องส่ง credentials

## Request และ response examples

`POST /api/v1/users` และ `PUT /api/v1/users/:id` ใช้ JSON body เดียวกัน:

```json
{
  "name": "สมชาย",
  "lastname": "ตัวอย่าง",
  "birthday": "1995-04-20",
  "email": "somchai@example.com",
  "tel": "0812345678"
}
```

Client ไม่ต้องส่ง `employee_id` หรือ `age`. `employee_id` สร้างอัตโนมัติ และ age คำนวณจาก birthday. Create ตอบ `201`:

```json
{
  "message": "employee created",
  "data": {
    "employee_id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "สมชาย",
    "lastname": "ตัวอย่าง",
    "birthday": "1995-04-20",
    "age": 31,
    "email": "somchai@example.com",
    "tel": "0812345678"
  }
}
```

ค่า age ในตัวอย่างขึ้นกับวันที่เรียกจริง เมื่ออ่านข้อมูลหรือจัดกลุ่ม ระบบคำนวณ age ใหม่จาก birthday. List/search ตอบ `{ "data": [...], "count": 25 }`; ไม่พบผลค้นหาตอบ `data: []` และ `count: 0`. Delete สำเร็จตอบ `204` ไม่มี body. ข้อมูลผิดตอบ `400` เช่น:

```json
{"error":"validation failed","details":{"name":"required; use Thai or English letters and spaces only"}}
```

UUID ผิดรูปแบบตอบ `400`; ไม่มี record ตอบ `404`; unique constraint ซ้ำตอบ `409`; database error ตอบ `500` โดยไม่เปิดเผยรายละเอียดภายใน

## Validation rules

| Field | Rule |
|---|---|
| `name`, `lastname` | ต้องมีค่า; ตัวอักษรไทย/อังกฤษ รวมสระและวรรณยุกต์ไทย; ช่องว่างระหว่างคำ; ไม่มีตัวเลขหรืออักขระพิเศษ |
| `birthday` | ต้องมีค่า, `YYYY-MM-DD`, เป็นวันที่จริง, ไม่ใช่อนาคต |
| `email` | ต้องมีค่าและเป็น email format ที่ถูกต้อง |
| `tel` | string ตัวเลข 9–15 หลัก อาจขึ้นต้นด้วย `+`; รองรับเลข 0 นำหน้า |

PUT ตรวจข้อมูลที่ส่งด้วยกฎเดียวกันและใช้ body ครบทุกช่อง

## Generation age ranges

| JSON key | Age |
|---|---|
| `gen_z` | น้อยกว่า 24 |
| `gen_y` | 24–41 |
| `gen_x` | 42–56 |
| `baby_boomer` | 57–75 |
| `gi_generation` | มากกว่า 75 |

`GET /api/v1/users/generations` คืน `{ "data": { "gen_z": [], ... }, "count": 25 }` โดยมีทั้ง 5 key เสมอ และแต่ละกลุ่มเป็น array แม้ไม่มีข้อมูล

## Search examples

- `GET /api/v1/users/search/Somchai`
- `GET /api/v1/users/search/สมชาย`
- `GET /api/v1/users/search/62832`

Keyword รับทาง **path parameter** ค่าเดียว ค้นหา partial match แบบ OR ใน `employee_id`, `name`, `lastname` ด้วย parameterized GORM query; URL encode keyword ภาษาไทยตามปกติ

## Tests และ Postman

```powershell
$env:CGO_ENABLED = '0'
go vet ./...
go test ./...
go build ./...
```

Import [Postman Collection](docs/postman_collection.json) แล้วเลือก request ที่ต้องการ Collection ตั้ง `baseUrl` เป็น `http://localhost:3000/api/v1`, `employeeId` เป็น UUID ที่ต้องการ และ `keyword` เป็นคำค้นหา ตัวอย่าง POST/PUT/DELETE ตั้ง Basic Auth ไว้ใน collection; GET ใช้ `noauth` สามารถเปลี่ยนตัวแปร username/password ได้ใน Postman

โค้ดและเอกสารพร้อมสำหรับ commit ไป GitHub; ไฟล์ `.env` และฐานข้อมูลจริงถูก ignore โดย Git
