# Full-stack TaskFlow — Hướng dẫn + toàn bộ source code

> Tutorial này đi kèm source code đầy đủ, gồm README hướng dẫn chạy local/deploy và nội dung của từng file.

# TaskFlow — Full-stack React + Express.js + PostgreSQL

Ứng dụng quản lý công việc dùng React 19 + Vite 7, Express 5 (Controller → Service → Repository), PostgreSQL 17. Đây là ứng dụng demo **không có authentication**: mọi người dùng truy cập API công khai đều có thể xem/sửa/xóa toàn bộ task. Không lưu dữ liệu nhạy cảm.

## 1. Tính năng

- Tạo/sửa/xóa task; đánh dấu hoàn thành
- Title, description, priority (low/medium/high), due date, status
- Tìm kiếm, lọc trạng thái và ưu tiên, phân trang
- Thống kê tổng task, đang làm, hoàn thành, quá hạn
- Validation bằng Zod, lỗi HTTP theo cấu trúc JSON, query có tham số chống SQL injection
- PostgreSQL migrations và Supertest integration tests
- Vite proxy khi phát triển local và biến môi trường cho deployment

## 2. Kiến trúc

```text
Browser (React + Vite :5173)
        | HTTP JSON /api (Vite proxy ở local)
        v
Express API (:4000) → Router → Validate → Controller → Service → Repository
                                                        |
                                                        v
                                                   PostgreSQL :5433
```

Cấu trúc thư mục:

```text
taskflow-fullstack/
├── compose.yaml
├── db/init/01-test-db.sql
├── api/
│   ├── .env.example, .env.test.example, package.json
│   ├── db/migrations/001_create_tasks.sql
│   ├── scripts/migrate.js
│   ├── tests/tasks.test.js
│   └── src/
│       ├── app.js, server.js
│       ├── config/db.js
│       ├── controllers/task.controller.js
│       ├── services/task.service.js
│       ├── repositories/task.repository.js
│       ├── routes/task.routes.js
│       ├── middleware/error.js, validate.js
│       └── schemas/task.schema.js
└── web/
    ├── .env.example, package.json, vite.config.js, index.html
    ├── public/favicon.svg
    └── src/
        ├── main.jsx, App.jsx, styles.css
        ├── components/TaskForm.jsx, TaskCard.jsx
        └── lib/api.js
```

## 3. Yêu cầu môi trường

- Node.js **22.12+** (hoặc phiên bản hỗ trợ tương đương)
- npm
- Docker + Docker Compose (hoặc PostgreSQL 17 cài trực tiếp)
- Git cho deployment

Kiểm tra:

```bash
node -v
npm -v
docker compose version
```

## 4. Chạy local từ đầu (3 terminal)

### Terminal 1 — PostgreSQL

Từ **thư mục gốc** `taskflow-fullstack`:

```bash
docker compose up -d
docker compose ps
```

Docker tải PostgreSQL 17 và tạo volume `pgdata`, database `task_manager`. Script `db/init/01-test-db.sql` tạo database `task_manager_test` trong lần đầu khởi tạo **volume mới**. Database được publish ra port `localhost:5433` để tránh đụng với PostgreSQL có sẵn ở `5432`.

**Không dùng tài khoản/mật khẩu mẫu của Docker Compose khi triển khai Internet.**

### Terminal 2 — Express backend

```bash
cd api
cp .env.example .env
cp .env.test.example .env.test
npm install
npm run db:migrate
npm run db:migrate:test
npm run dev
```

Windows PowerShell thay `cp` bằng `Copy-Item` (cú pháp `cp` cũng thường là alias có sẵn).

Kiểm tra backend:

```bash
curl http://localhost:4000/health
curl http://localhost:4000/ready
curl http://localhost:4000/api/tasks
```

- `/health`: HTTP server hoạt động.
- `/ready`: backend truy cập được PostgreSQL.
- `/api/tasks`: JSON list từ database.

### Terminal 3 — React frontend

```bash
cd web
cp .env.example .env
npm install
npm run dev
```

Mở **http://localhost:5173**. Trong local, `VITE_API_URL=` (rỗng); Vite proxy các request `/api` về `http://localhost:4000`. Điều này tránh phải hardcode backend localhost URL vào code React và giúp giữ cách gọi `/api/...` nhất quán.

> Nếu `.env` của web có `VITE_API_URL=https://...`, sẽ gọi remote API thay vì proxy local. Khi chỉnh `.env`, restart Vite.

## 5. Khám phá nghiệp vụ

### API endpoints

| Method | Path | Chức năng |
|---|---|---|
| GET | `/health` | Liveness |
| GET | `/ready` | Readiness / PostgreSQL |
| GET | `/api/tasks` | Danh sách, tìm kiếm, lọc, phân trang |
| GET | `/api/tasks/stats` | Thống kê toàn bộ task |
| GET | `/api/tasks/:id` | Xem 1 task |
| POST | `/api/tasks` | Tạo task |
| PATCH | `/api/tasks/:id` | Cập nhật một phần |
| DELETE | `/api/tasks/:id` | Xóa task |

Ví dụ tạo:

```bash
curl -i -X POST http://localhost:4000/api/tasks \
  -H 'Content-Type: application/json' \
  -d '{"title":"Học React","description":"Hoàn thiện UI","priority":"high","due_date":"2027-01-31"}'
```

Lọc:

```bash
curl 'http://localhost:4000/api/tasks?q=React&status=active&priority=high&page=1&limit=8'
```

Cập nhật:

```bash
curl -i -X PATCH http://localhost:4000/api/tasks/1 \
  -H 'Content-Type: application/json' -d '{"completed":true}'
```

Xóa:

```bash
curl -i -X DELETE http://localhost:4000/api/tasks/1
```

**Dữ liệu Task:**

```json
{
  "id": 1,
  "title": "Học React",
  "description": "Hoàn thiện UI",
  "priority": "high",
  "due_date": "2027-01-31",
  "completed": false,
  "created_at": "2026-11-09T19:00:00.000Z",
  "updated_at": "2026-11-09T19:00:00.000Z"
}
```

`created_at/updated_at` là timestamp minh họa; thực tế PostgreSQL sinh theo thời gian của server.

**HTTP status:** 200 GET/PATCH; 201 POST; 204 DELETE; 400 validation; 404 missing resource; 500 unexpected server errors.

## 6. Mô hình 3 tầng — học theo thứ tự

1. `api/src/routes/task.routes.js`: ghép method+path với middleware và controller. `GET /stats` đặt trước `GET /:id`.
2. `api/src/middleware/validate.js` và `api/src/schemas/task.schema.js`: parse query/body/params, trả 400 khi lỗi. Giá trị đã parse được lưu trong `res.locals.validated`.
3. `api/src/controllers/task.controller.js`: gọi Service, quyết định status HTTP và JSON.
4. `api/src/services/task.service.js`: nơi xử lý nghiệp vụ, ném `AppError(404, ...)` khi không tìm thấy task.
5. `api/src/repositories/task.repository.js`: SQL, tìm kiếm, lọc, thống kê, phân trang, CRUD và bind parameters (`$1`, `$2`, ...).
6. `api/src/config/db.js`: PostgreSQL connection pool dùng chung.
7. `api/src/middleware/error.js`: lỗi tập trung, không tiết lộ nội bộ cho client.

**Data flow tạo task:** Form → `taskApi.create()` → POST `/api/tasks` → Router → Zod → Controller → Service → Repository → `INSERT ... RETURNING` → HTTP 201 → React reload danh sách + số liệu.

**Data flow tìm kiếm:** React debounce 300ms → GET `/api/tasks?q=...` → query đã validate → Repository `POSITION(LOWER($1) IN LOWER(title))` → JSON pagination → React render.

Repository chỉ nhận các giá trị được validate; tên cột cập nhật chọn từ whitelist hardcode. Không bao giờ nối trực tiếp input từ người dùng vào câu SQL.

## 7. Kiểm thử backend

```bash
cd api
npm run db:migrate:test
npm test
```

Các test tạo/xóa dữ liệu trong **task_manager_test** và chạy `TRUNCATE tasks RESTART IDENTITY` trước mỗi test. Không chạy bằng production database. Script bảo vệ bằng `NODE_ENV=test` và quy tắc database phải có tên kết thúc `_test`.

Nếu dùng npm cài dependency lần đầu, nên commit `api/package-lock.json` và `web/package-lock.json` vào Git để build tái lập được. Các lockfile không kèm sẵn trong archive vì công cụ tạo project không truy cập npm registry.

## 8. Build React trước khi deploy

```bash
cd web
npm run build
npm run preview
```

Mặc định Vite tạo thư mục `web/dist`. Đó là artifact để deploy Static Site, không phải Express server. Preview chỉ phục vụ kiểm tra local; không dùng làm production web server.

## 9. Deploy lên Render: PostgreSQL + Express + React

### Chuẩn bị GitHub

Tạo repo GitHub mới, rồi từ thư mục root:

```bash
git init
git add .
git commit -m 'Build TaskFlow full-stack app'
git branch -M main
git remote add origin https://github.com/YOUR_USERNAME/taskflow-fullstack.git
git push -u origin main
```

Đảm bảo `.env` và `.env.test` không có trong commit (`.gitignore`).

### Bước A — Tạo PostgreSQL

1. Truy cập https://dashboard.render.com/.
2. Chọn **New → Postgres**, đặt tên database, chọn region và plan phù hợp. Giá / giới hạn plan có thể thay đổi; kiểm tra dashboard trước khi tạo.
3. Sau khi database sẵn sàng, lấy **Internal Database URL**. Nên đặt backend cùng region và cùng workspace/account để dùng mạng nội bộ.

### Bước B — Deploy Express API dưới dạng Web Service

1. Chọn **New → Web Service**, kết nối GitHub repo.
2. Chọn `Root Directory = api` (monorepo).
3. Runtime/Language: `Node`.
4. Build Command: `npm install` (hoặc `npm ci` sau khi commit lockfile).
5. Start Command: `npm run start:deploy`.
6. Đặt environment variables:

   | Key | Value |
   |---|---|
   | `NODE_ENV` | `production` |
   | `DATABASE_URL` | Internal Database URL từ bước A |
   | `FRONTEND_ORIGIN` | URL frontend Render từ bước C, ví dụ `https://your-taskflow.onrender.com` |

7. Nếu thiết lập health check path, đặt `/ready`. API phải listen trên `0.0.0.0` và port `process.env.PORT`; source đã hỗ trợ.
8. Deploy. Sau khi thành công, kiểm tra `https://your-api.onrender.com/health` và `/ready`.

**Migrations khi deploy:** `start:deploy` chạy migration trước khi start Express. `scripts/migrate.js` lưu danh sách migration đã áp dụng trong bảng `schema_migrations` và dùng PostgreSQL advisory transaction lock để tránh 2 tiến trình chạy migration chồng chéo. Đây là lựa chọn tiện lợi cho demo hoặc một instance; với hệ thống production lớn nên đưa migration vào pre-deploy job/pipeline riêng, bảo đảm migration backward-compatible trước khi roll out nhiều instance.

> `FRONTEND_ORIGIN` cần chính xác origin frontend (không có dấu `/` ở cuối). Bạn có thể cập nhật biến này và redeploy sau khi đã tạo frontend. Nếu muốn thử backend bằng curl/Postman thì không cần CORS, vì chúng không thực thi chính sách CORS của trình duyệt.

### Bước C — Deploy React dưới dạng Static Site

1. Chọn **New → Static Site**, kết nối cùng GitHub repo.
2. `Root Directory = web`.
3. Build Command: `npm install && npm run build` (hoặc `npm ci && npm run build` nếu có lockfile).
4. Publish Directory: `dist`.
5. Environment variable:

   | Key | Value |
   |---|---|
   | `VITE_API_URL` | `https://your-api.onrender.com` |

6. Deploy. Mở URL `https://your-taskflow.onrender.com` được Render cung cấp.
7. Quay lại backend và sửa `FRONTEND_ORIGIN` thành URL vừa được cấp, sau đó redeploy backend nếu trước đó dùng giá trị tạm.
8. Làm thử: Tạo task → reload trang → task vẫn còn, đánh dấu hoàn thành → cập nhật số liệu → xóa.

**Chú ý:** `VITE_API_URL` phải là URL API công khai (frontend chạy trong trình duyệt); `DATABASE_URL` là URL private giữa backend và PostgreSQL, không đưa vào frontend. Các biến `VITE_*` sẽ được đưa vào bundle build React nên tuyệt đối không chứa password, secret hoặc credential.

### Luồng sau deployment

```text
https://taskflow-frontend.onrender.com (Static Site)
      | HTTPS fetch('/api/tasks') qua VITE_API_URL
      v
https://taskflow-backend.onrender.com (Web Service)
      | DATABASE_URL (nội bộ)
      v
Render PostgreSQL (cùng region)
```

**Lưu ý phí và hiệu năng:** các plan, giới hạn tài nguyên và chính sách free/spin-down có thể thay đổi. Kiểm tra bảng giá Render hiện hành trước khi sử dụng. Các web service free (nếu đang khả dụng) có thể ngủ khi không có request và cần khởi động lại ở lần gọi đầu.

## 10. Lỗi thường gặp

| Hiện tượng | Nguyên nhân / cách xử lý |
|---|---|
| `ECONNREFUSED :5433` | Docker chưa chạy; thử `docker compose ps` |
| `relation "tasks" does not exist` | Chưa chạy `npm run db:migrate` |
| `database "task_manager_test" does not exist` | Volume cũ không chạy init SQL; tự tạo test DB hoặc dùng volume mới (không xóa dữ liệu quan trọng) |
| React hiển thị `Không thể kết nối API` | Backend chưa bật, sai `VITE_API_URL` hoặc deploy backend chưa xong |
| CORS error trên trình duyệt | `FRONTEND_ORIGIN` khác URL frontend (scheme/domain/port), hoặc backend đang lỗi |
| `404 /api/tasks` ở frontend | Vite proxy chưa chạy, sai path, thiếu `/api` |
| Render không phát hiện port | Phải bind `0.0.0.0:$PORT`; source đã thực hiện |
| Bấm lưu nhận 400 | Kiểm tra title 1–120 ký tự, priority, YYYY-MM-DD, kiểu boolean |
| Migrations thất bại | Kiểm tra `DATABASE_URL`, quyền `CREATE TABLE`, log deploy |

## 11. Những điều cần bổ sung trước khi dùng production nhiều người

Ứng dụng hiện tại **không có xác thực**. CORS **không** bảo vệ API trước truy cập trái phép. Trước khi lưu dữ liệu người dùng thật cần bổ sung: đăng ký/đăng nhập, JWT hoặc session được bảo vệ, phân quyền theo `user_id`, rate limiting, logging và monitoring, bảo vệ dữ liệu cá nhân, migration tương thích ngược, bảo trì backup/restore và test riêng cho quyền truy cập.

## 12. Nguồn tham khảo chính thức

- https://vite.dev/guide/
- https://vite.dev/guide/static-deploy.html
- https://expressjs.com/en/5x/guide/error-handling/
- https://node-postgres.com/features/queries
- https://node-postgres.com/features/pooling
- https://render.com/docs/multi-service-architecture
- https://render.com/docs/postgresql-creating-connecting
- https://render.com/docs/web-services



---

# Source code đầy đủ từng file


## `.gitignore`


```gitignore
**/node_modules/
**/.env
**/.env.local
**/.env.test
**/.env.production
**/dist/
**/coverage/
.DS_Store
*.log
```


## `api/.env.example`


```dotenv
NODE_ENV=development
PORT=4000
DATABASE_URL=postgresql://taskuser:taskpass_local_only@localhost:5433/task_manager
FRONTEND_ORIGIN=http://localhost:5173
```


## `api/.env.test.example`


```dotenv
NODE_ENV=test
PORT=4001
DATABASE_URL=postgresql://taskuser:taskpass_local_only@localhost:5433/task_manager_test
FRONTEND_ORIGIN=http://localhost:5173
```


## `api/db/migrations/001_create_tasks.sql`


```sql
CREATE TABLE tasks (
  id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  title VARCHAR(120) NOT NULL CHECK (length(btrim(title)) > 0),
  description VARCHAR(2000) NOT NULL DEFAULT '',
  priority VARCHAR(8) NOT NULL DEFAULT 'medium'
    CHECK (priority IN ('low', 'medium', 'high')),
  due_date DATE,
  completed BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tasks_created ON tasks (created_at DESC, id DESC);
CREATE INDEX idx_tasks_filters ON tasks (completed, priority);
```


## `api/package.json`


```json
{
  "name": "taskflow-api",
  "version": "1.0.0",
  "private": true,
  "type": "module",
  "engines": {"node": ">=22.12.0"},
  "scripts": {
    "dev": "node --watch src/server.js",
    "start": "node src/server.js",
    "start:deploy": "node scripts/migrate.js && node src/server.js",
    "db:migrate": "node scripts/migrate.js",
    "db:migrate:test": "node --env-file=.env.test scripts/migrate.js",
    "test": "node --env-file=.env.test --test"
  },
  "dependencies": {
    "cors": "^2.8.5",
    "dotenv": "^16.6.0",
    "express": "^5.1.0",
    "helmet": "^8.1.0",
    "pg": "^8.16.0",
    "zod": "^4.0.0"
  },
  "devDependencies": {
    "supertest": "^7.1.0"
  }
}
```


## `api/scripts/migrate.js`


```javascript
import 'dotenv/config';
import { readFile, readdir } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';
import { dirname, join } from 'node:path';
import { pool } from '../src/config/db.js';

const directory = join(dirname(fileURLToPath(import.meta.url)), '../db/migrations');
const files = (await readdir(directory)).filter((file) => file.endsWith('.sql')).sort();
const client = await pool.connect();

try {
  await client.query('BEGIN');
  // Serialize competing migrations from multiple deploys/processes.
  await client.query('SELECT pg_advisory_xact_lock(902101)');
  await client.query(`CREATE TABLE IF NOT EXISTS schema_migrations (
    filename TEXT PRIMARY KEY,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
  )`);

  for (const file of files) {
    const exists = await client.query(
      'SELECT 1 FROM schema_migrations WHERE filename = $1', [file]
    );
    if (exists.rowCount) {
      console.log(`Already applied: ${file}`);
      continue;
    }
    await client.query(await readFile(join(directory, file), 'utf8'));
    await client.query('INSERT INTO schema_migrations (filename) VALUES ($1)', [file]);
    console.log(`Applied: ${file}`);
  }
  await client.query('COMMIT');
} catch (error) {
  await client.query('ROLLBACK');
  console.error('Migration failed:', error);
  process.exitCode = 1;
} finally {
  client.release();
  await pool.end();
}
```


## `api/src/app.js`


```javascript
import 'dotenv/config';
import express from 'express';
import cors from 'cors';
import helmet from 'helmet';
import { pool } from './config/db.js';
import taskRoutes from './routes/task.routes.js';
import { AppError, notFound, errorHandler } from './middleware/error.js';

const app = express();
const origins = (process.env.FRONTEND_ORIGIN || 'http://localhost:5173')
  .split(',').map((value) => value.trim()).filter(Boolean);

app.disable('x-powered-by');
app.use(helmet());
app.use(cors({
  origin(origin, callback) {
    if (!origin || origins.includes(origin)) callback(null, true);
    else callback(new AppError(403, 'Origin not allowed'));
  }
}));
app.use(express.json({ limit: '100kb' }));
app.get('/health', (req, res) => res.json({ status: 'ok' }));
app.get('/ready', async (req, res) => {
  await pool.query('SELECT 1');
  res.json({ status: 'ready' });
});
app.use('/api/tasks', taskRoutes);
app.use(notFound);
app.use(errorHandler);

export default app;
```


## `api/src/config/db.js`


```javascript
import 'dotenv/config';
import { Pool } from 'pg';

if (!process.env.DATABASE_URL) {
  throw new Error('DATABASE_URL is required');
}

// Tests must never touch the development or production database.
if (process.env.NODE_ENV === 'test') {
  const database = decodeURIComponent(new URL(process.env.DATABASE_URL).pathname.slice(1));
  if (!database.endsWith('_test')) {
    throw new Error('A test database must have a name ending with _test');
  }
}

export const pool = new Pool({
  connectionString: process.env.DATABASE_URL,
  max: 10,
  connectionTimeoutMillis: 5000,
  idleTimeoutMillis: 30000
});

pool.on('error', (error) => {
  console.error('Unexpected idle PostgreSQL client error', error);
});
```


## `api/src/controllers/task.controller.js`


```javascript
import { taskService } from '../services/task.service.js';

export const taskController = {
  async list(req, res) {
    const { items, pagination } = await taskService.list(res.locals.validated.query);
    res.json({ success: true, data: items, pagination });
  },
  async stats(req, res) {
    res.json({ success: true, data: await taskService.stats() });
  },
  async get(req, res) {
    const task = await taskService.get(res.locals.validated.params.id);
    res.json({ success: true, data: task });
  },
  async create(req, res) {
    const task = await taskService.create(res.locals.validated.body);
    res.location(`/api/tasks/${task.id}`).status(201).json({ success: true, data: task });
  },
  async update(req, res) {
    const task = await taskService.update(
      res.locals.validated.params.id, res.locals.validated.body
    );
    res.json({ success: true, data: task });
  },
  async remove(req, res) {
    await taskService.remove(res.locals.validated.params.id);
    res.status(204).end();
  }
};
```


## `api/src/middleware/error.js`


```javascript
export class AppError extends Error {
  constructor(status, message, details) {
    super(message);
    this.status = status;
    this.details = details;
  }
}

export function notFound(req, res) {
  res.status(404).json({ success: false, error: { message: 'Route not found' } });
}

export function errorHandler(err, req, res, next) {
  if (res.headersSent) return next(err);
  const status = Number.isInteger(err.status) && err.status >= 400 && err.status < 600
    ? err.status : 500;
  if (status >= 500) console.error(err);
  const message = err.type === 'entity.parse.failed'
    ? 'Invalid JSON body'
    : status >= 500 ? 'Internal server error' : err.message;
  res.status(status).json({
    success: false,
    error: {
      message,
      ...(status < 500 && err.details ? { details: err.details } : {})
    }
  });
}
```


## `api/src/middleware/validate.js`


```javascript
import { AppError } from './error.js';

export const validate = (schema, location = 'body') => (req, res, next) => {
  const result = schema.safeParse(req[location]);
  if (!result.success) {
    throw new AppError(400, 'Invalid request data', result.error.issues.map((issue) => ({
      field: issue.path.join('.') || location,
      message: issue.message
    })));
  }
  res.locals.validated ??= {};
  res.locals.validated[location] = result.data;
  next();
};
```


## `api/src/repositories/task.repository.js`


```javascript
import { pool } from '../config/db.js';

const columns = `id, title, description, priority, due_date::text AS due_date,
                 completed, created_at, updated_at`;

const filterSql = `
  WHERE ($1::text IS NULL
    OR POSITION(LOWER($1) IN LOWER(title)) > 0
    OR POSITION(LOWER($1) IN LOWER(description)) > 0)
    AND ($2::text = 'all'
      OR ($2::text = 'done' AND completed = TRUE)
      OR ($2::text = 'active' AND completed = FALSE))
    AND ($3::text = 'all' OR priority = $3)
`;

export const taskRepository = {
  async findMany({ q, status, priority, page, limit }) {
    const filters = [q || null, status, priority];
    const offset = (page - 1) * limit;
    const [itemsResult, countResult] = await Promise.all([
      pool.query(
        `SELECT ${columns} FROM tasks ${filterSql}
         ORDER BY created_at DESC, id DESC LIMIT $4 OFFSET $5`,
        [...filters, limit, offset]
      ),
      pool.query(`SELECT COUNT(*)::int AS total FROM tasks ${filterSql}`, filters)
    ]);
    return { items: itemsResult.rows, total: countResult.rows[0].total };
  },

  async stats() {
    const result = await pool.query(`
      SELECT COUNT(*)::int AS total,
        COUNT(*) FILTER (WHERE completed = TRUE)::int AS done,
        COUNT(*) FILTER (WHERE completed = FALSE)::int AS active,
        COUNT(*) FILTER (WHERE completed = FALSE AND due_date < CURRENT_DATE)::int AS overdue
      FROM tasks
    `);
    return result.rows[0];
  },

  async findById(id) {
    const result = await pool.query(`SELECT ${columns} FROM tasks WHERE id = $1`, [id]);
    return result.rows[0] ?? null;
  },

  async create({ title, description, priority, due_date, completed }) {
    const result = await pool.query(
      `INSERT INTO tasks (title, description, priority, due_date, completed)
       VALUES ($1, $2, $3, $4, $5) RETURNING ${columns}`,
      [title, description, priority, due_date, completed]
    );
    return result.rows[0];
  },

  async update(id, changes) {
    const permitted = ['title', 'description', 'priority', 'due_date', 'completed'];
    const entries = permitted.filter((key) => Object.hasOwn(changes, key));
    const assignments = entries.map((key, i) => `${key} = $${i + 1}`);
    assignments.push('updated_at = NOW()');
    const values = [...entries.map((key) => changes[key]), id];
    const result = await pool.query(
      `UPDATE tasks SET ${assignments.join(', ')}
       WHERE id = $${values.length} RETURNING ${columns}`,
      values
    );
    return result.rows[0] ?? null;
  },

  async remove(id) {
    const result = await pool.query('DELETE FROM tasks WHERE id = $1', [id]);
    return result.rowCount > 0;
  }
};
```


## `api/src/routes/task.routes.js`


```javascript
import { Router } from 'express';
import { taskController } from '../controllers/task.controller.js';
import { validate } from '../middleware/validate.js';
import { paramsSchema, listSchema, createSchema, updateSchema } from '../schemas/task.schema.js';

const router = Router();
router.get('/', validate(listSchema, 'query'), taskController.list);
router.get('/stats', taskController.stats); // Before /:id
router.get('/:id', validate(paramsSchema, 'params'), taskController.get);
router.post('/', validate(createSchema), taskController.create);
router.patch('/:id', validate(paramsSchema, 'params'), validate(updateSchema), taskController.update);
router.delete('/:id', validate(paramsSchema, 'params'), taskController.remove);
export default router;
```


## `api/src/schemas/task.schema.js`


```javascript
import { z } from 'zod';

const idSchema = z.string().regex(/^[1-9]\d*$/).transform(Number)
  .pipe(z.int().min(1).max(2147483647));

const dateSchema = z.string().regex(/^\d{4}-\d{2}-\d{2}$/)
  .refine((value) => {
    const date = new Date(`${value}T00:00:00.000Z`);
    return !Number.isNaN(date.getTime()) && date.toISOString().slice(0, 10) === value;
  }, 'Date must be valid and formatted YYYY-MM-DD');

const taskFields = {
  title: z.string().trim().min(1).max(120),
  description: z.string().trim().max(2000),
  priority: z.enum(['low', 'medium', 'high']),
  due_date: dateSchema.nullable(),
  completed: z.boolean()
};

const positiveIntQuery = (max, defaultValue) => z.coerce.number()
  .int().min(1).max(max).default(defaultValue);

export const paramsSchema = z.strictObject({ id: idSchema });
export const listSchema = z.strictObject({
  q: z.string().trim().max(100).optional(),
  status: z.enum(['all', 'active', 'done']).default('all'),
  priority: z.enum(['all', 'low', 'medium', 'high']).default('all'),
  page: positiveIntQuery(100000, 1),
  limit: positiveIntQuery(50, 8)
});
export const createSchema = z.strictObject({
  title: taskFields.title,
  description: taskFields.description.default(''),
  priority: taskFields.priority.default('medium'),
  due_date: taskFields.due_date.default(null),
  completed: taskFields.completed.default(false)
});
export const updateSchema = z.strictObject({
  title: taskFields.title.optional(),
  description: taskFields.description.optional(),
  priority: taskFields.priority.optional(),
  due_date: taskFields.due_date.optional(),
  completed: taskFields.completed.optional()
}).refine((value) => Object.keys(value).length > 0, 'At least one field is required');
```


## `api/src/server.js`


```javascript
import app from './app.js';
import { pool } from './config/db.js';

const port = Number(process.env.PORT || 4000);
const server = app.listen(port, '0.0.0.0', () => {
  console.log(`TaskFlow API listening on port ${port}`);
});

async function shutdown(signal) {
  console.log(`${signal} received, shutting down`);
  server.close(async () => {
    try { await pool.end(); process.exit(0); }
    catch (error) { console.error(error); process.exit(1); }
  });
  // Avoid hanging indefinitely on existing connections.
  server.closeIdleConnections?.();
}
process.on('SIGTERM', () => shutdown('SIGTERM'));
process.on('SIGINT', () => shutdown('SIGINT'));
```


## `api/src/services/task.service.js`


```javascript
import { taskRepository } from '../repositories/task.repository.js';
import { AppError } from '../middleware/error.js';

export const taskService = {
  async list(query) {
    const { items, total } = await taskRepository.findMany(query);
    return {
      items,
      pagination: {
        page: query.page,
        limit: query.limit,
        total,
        totalPages: Math.ceil(total / query.limit)
      }
    };
  },
  stats: () => taskRepository.stats(),
  async get(id) {
    const task = await taskRepository.findById(id);
    if (!task) throw new AppError(404, 'Task not found');
    return task;
  },
  create: (body) => taskRepository.create(body),
  async update(id, body) {
    const task = await taskRepository.update(id, body);
    if (!task) throw new AppError(404, 'Task not found');
    return task;
  },
  async remove(id) {
    if (!await taskRepository.remove(id)) throw new AppError(404, 'Task not found');
  }
};
```


## `api/tests/tasks.test.js`


```javascript
import { beforeEach, after, test } from 'node:test';
import assert from 'node:assert/strict';
import request from 'supertest';
import app from '../src/app.js';
import { pool } from '../src/config/db.js';

// This test suite requires task_manager_test, migrated with db:migrate:test.
beforeEach(async () => {
  await pool.query('TRUNCATE TABLE tasks RESTART IDENTITY');
});
after(async () => { await pool.end(); });

test('health/readiness endpoints', async () => {
  await request(app).get('/health').expect(200);
  await request(app).get('/ready').expect(200);
});

test('create, get, patch and delete a task', async () => {
  const created = await request(app).post('/api/tasks').send({
    title: 'Write tests', description: 'End to end API', priority: 'high', due_date: '2027-01-31'
  }).expect(201);
  const id = created.body.data.id;
  assert.equal(created.body.data.title, 'Write tests');
  assert.equal(created.body.data.due_date, '2027-01-31');
  const fetched = await request(app).get(`/api/tasks/${id}`).expect(200);
  assert.equal(fetched.body.data.priority, 'high');
  const updated = await request(app).patch(`/api/tasks/${id}`)
    .send({ completed: true }).expect(200);
  assert.equal(updated.body.data.completed, true);
  await request(app).delete(`/api/tasks/${id}`).expect(204);
  await request(app).get(`/api/tasks/${id}`).expect(404);
});

test('search, filter, pagination and statistics', async () => {
  await request(app).post('/api/tasks').send({ title: 'Learn React', completed: true }).expect(201);
  await request(app).post('/api/tasks').send({ title: 'Learn SQL', priority: 'high' }).expect(201);
  const result = await request(app).get('/api/tasks?q=Learn&status=active&priority=high&page=1&limit=2').expect(200);
  assert.equal(result.body.pagination.total, 1);
  assert.equal(result.body.data[0].title, 'Learn SQL');
  const stats = await request(app).get('/api/tasks/stats').expect(200);
  assert.equal(stats.body.data.total, 2);
  assert.equal(stats.body.data.done, 1);
  assert.equal(stats.body.data.active, 1);
});

test('invalid requests are rejected', async () => {
  await request(app).post('/api/tasks').send({ title: '' }).expect(400);
  await request(app).post('/api/tasks').send({ title: 'X', priority: 'urgent' }).expect(400);
  await request(app).post('/api/tasks').send({ title: 'X', due_date: '2027-02-30' }).expect(400);
  await request(app).get('/api/tasks?page=0').expect(400);
  await request(app).get('/api/tasks/abc').expect(400);
  await request(app).patch('/api/tasks/1').send({}).expect(400);
});
```


## `compose.yaml`


```yaml
services:
  postgres:
    image: postgres:17-alpine
    restart: unless-stopped
    environment:
      POSTGRES_USER: taskuser
      POSTGRES_PASSWORD: taskpass_local_only
      POSTGRES_DB: task_manager
    ports:
      - "5433:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data
      - ./db/init:/docker-entrypoint-initdb.d:ro
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U taskuser -d task_manager"]
      interval: 5s
      timeout: 5s
      retries: 12
volumes:
  pgdata:
```


## `db/init/01-test-db.sql`


```sql
CREATE DATABASE task_manager_test;
```


## `web/.env.example`


```dotenv
# Leave blank for local development: Vite proxies /api to Express on port 4000.
# Production Render Static Site: set VITE_API_URL=https://your-api.onrender.com
VITE_API_URL=
```


## `web/index.html`


```html
<!doctype html>
<html lang="vi">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <meta name="theme-color" content="#f5f7fc" />
    <meta name="description" content="TaskFlow - quản lý công việc với React, Express và PostgreSQL" />
    <link rel="icon" href="/favicon.svg" type="image/svg+xml" />
    <title>TaskFlow · Quản lý công việc</title>
  </head>
  <body>
    <div id="root"></div>
    <script type="module" src="/src/main.jsx"></script>
  </body>
</html>
```


## `web/package.json`


```json
{
  "name": "taskflow-web",
  "version": "1.0.0",
  "private": true,
  "type": "module",
  "engines": {"node": ">=22.12.0"},
  "scripts": {
    "dev": "vite --host 0.0.0.0",
    "build": "vite build",
    "preview": "vite preview --host 0.0.0.0"
  },
  "dependencies": {
    "react": "^19.1.0",
    "react-dom": "^19.1.0"
  },
  "devDependencies": {
    "@vitejs/plugin-react": "^5.0.0",
    "vite": "^7.1.0"
  }
}
```


## `web/public/favicon.svg`


```xml
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64">
  <rect width="64" height="64" rx="16" fill="#365cf5"/>
  <path d="m17 32 10 10 21-22" fill="none" stroke="white" stroke-width="7" stroke-linecap="round" stroke-linejoin="round"/>
</svg>
```


## `web/src/App.jsx`


```jsx
import { useEffect, useState } from 'react';
import { taskApi } from './lib/api.js';
import TaskCard from './components/TaskCard.jsx';
import TaskForm from './components/TaskForm.jsx';

const initialPagination = { page: 1, limit: 8, total: 0, totalPages: 0 };

export default function App() {
  const [tasks, setTasks] = useState([]);
  const [stats, setStats] = useState({ total: 0, active: 0, done: 0, overdue: 0 });
  const [pagination, setPagination] = useState(initialPagination);
  const [search, setSearch] = useState('');
  const [debouncedSearch, setDebouncedSearch] = useState('');
  const [status, setStatus] = useState('all');
  const [priority, setPriority] = useState('all');
  const [page, setPage] = useState(1);
  const [refreshKey, setRefreshKey] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [modalOpen, setModalOpen] = useState(false);
  const [editingTask, setEditingTask] = useState(null);
  const [formError, setFormError] = useState('');
  const [saving, setSaving] = useState(false);
  const [busyId, setBusyId] = useState(null);

  useEffect(() => {
    const timer = setTimeout(() => setDebouncedSearch(search.trim()), 300);
    return () => clearTimeout(timer);
  }, [search]);

  useEffect(() => {
    let alive = true;
    setLoading(true);
    setError('');
    Promise.all([
      taskApi.list({ q: debouncedSearch, status, priority, page, limit: 8 }),
      taskApi.stats()
    ]).then(([listResult, statsResult]) => {
      if (!alive) return;
      setTasks(listResult.data);
      setPagination(listResult.pagination);
      setStats(statsResult.data);
    }).catch((err) => {
      if (alive) setError(err.message);
    }).finally(() => {
      if (alive) setLoading(false);
    });
    return () => { alive = false; };
  }, [debouncedSearch, status, priority, page, refreshKey]);

  function changeSearch(value) { setSearch(value); setPage(1); }
  function changeStatus(value) { setStatus(value); setPage(1); }
  function changePriority(value) { setPriority(value); setPage(1); }
  function refresh() { setRefreshKey((key) => key + 1); }
  function openCreate() { setEditingTask(null); setFormError(''); setModalOpen(true); }
  function openEdit(task) { setEditingTask(task); setFormError(''); setModalOpen(true); }

  async function saveTask(payload) {
    setFormError('');
    setSaving(true);
    try {
      if (editingTask) await taskApi.update(editingTask.id, payload);
      else await taskApi.create(payload);
      setModalOpen(false);
      setPage(1);
      refresh();
    } catch (err) {
      setFormError(err.message);
    } finally {
      setSaving(false);
    }
  }

  async function toggleTask(task) {
    setBusyId(task.id);
    setError('');
    try {
      await taskApi.update(task.id, { completed: !task.completed });
      refresh();
    } catch (err) { setError(err.message); }
    finally { setBusyId(null); }
  }

  async function deleteTask(task) {
    if (!window.confirm(`Xóa công việc “${task.title}”? Thao tác này không thể hoàn tác.`)) return;
    setBusyId(task.id);
    setError('');
    try {
      await taskApi.remove(task.id);
      if (tasks.length === 1 && page > 1) setPage((value) => value - 1);
      else refresh();
    } catch (err) { setError(err.message); }
    finally { setBusyId(null); }
  }

  return (
    <div className="app-shell">
      <aside className="sidebar">
        <div className="brand"><div className="brand-icon">✓</div><span>Task<span className="brand-accent">Flow</span></span></div>
        <p className="nav-caption">WORKSPACE</p>
        <a className="nav-item active" href="#overview"><span>▦</span> Tổng quan</a>
        <a className="nav-item" href="#tasks"><span>☑</span> Công việc</a>
        <div className="sidebar-bottom">
          <div className="sidebar-tip"><strong>Focus on progress.</strong><p>Mỗi công việc hoàn thành là một bước tiến.</p></div>
          <span className="sidebar-version">React · Express · PostgreSQL</span>
        </div>
      </aside>
      <main className="main-content">
        <header className="topbar"><div className="mobile-brand">TaskFlow</div><div className="topbar-right"><span className="live-dot" /> Task Management Workspace <span className="avatar">TF</span></div></header>
        <div className="content-container">
          <div className="page-heading" id="overview">
            <div><span className="eyebrow">YOUR WORKSPACE</span><h1>Quản lý công việc<span className="heading-period">.</span></h1><p>Tập trung, sắp xếp ưu tiên và hoàn thành mục tiêu mỗi ngày.</p></div>
            <button className="button button-primary add-button" onClick={openCreate}><span aria-hidden="true">＋</span> Thêm công việc</button>
          </div>
          <section className="stats-grid" aria-label="Thống kê công việc">
            {[
              { label: 'Tổng công việc', value: stats.total, icon: '▦', tone: 'blue' },
              { label: 'Đang thực hiện', value: stats.active, icon: '◷', tone: 'amber' },
              { label: 'Đã hoàn thành', value: stats.done, icon: '✓', tone: 'green' },
              { label: 'Quá hạn', value: stats.overdue, icon: '!', tone: 'red' }
            ].map((item) => <div className="stat-card" key={item.label}>
              <div className={`stat-icon tone-${item.tone}`}>{item.icon}</div>
              <div><div className="stat-number">{item.value}</div><div className="stat-label">{item.label}</div></div>
            </div>)}
          </section>
          <section className="task-section" id="tasks">
            <div className="section-heading"><div><h2>Danh sách công việc</h2><p>{pagination.total} kết quả phù hợp</p></div><span className="section-badge">Tasks</span></div>
            <div className="filters">
              <div className="search-field"><span aria-hidden="true">⌕</span><input aria-label="Tìm kiếm công việc" placeholder="Tìm theo tên hoặc mô tả..." value={search} onChange={(event) => changeSearch(event.target.value)} /></div>
              <select aria-label="Lọc trạng thái" value={status} onChange={(event) => changeStatus(event.target.value)}>
                <option value="all">Mọi trạng thái</option><option value="active">Đang thực hiện</option><option value="done">Đã hoàn thành</option>
              </select>
              <select aria-label="Lọc mức ưu tiên" value={priority} onChange={(event) => changePriority(event.target.value)}>
                <option value="all">Mọi ưu tiên</option><option value="high">Ưu tiên cao</option><option value="medium">Trung bình</option><option value="low">Ưu tiên thấp</option>
              </select>
            </div>
            {error && <div className="notice error-notice" role="alert">{error} <button className="inline-retry" onClick={refresh}>Thử lại</button></div>}
            {loading ? <div className="empty-state" role="status">Đang tải dữ liệu từ PostgreSQL…</div>
              : !error && tasks.length === 0 ? <div className="empty-state"><div className="empty-icon">☑</div><h3>Chưa có công việc phù hợp</h3><p>Hãy tạo công việc mới hoặc thay đổi bộ lọc.</p><button className="button button-primary" onClick={openCreate}>＋ Tạo công việc</button></div>
              : <div className="task-list">{tasks.map((task) => <TaskCard key={task.id} task={task}
                busy={busyId === task.id} onToggle={toggleTask} onEdit={openEdit} onDelete={deleteTask} />)}</div>}
            <div className="pagination"><span>Trang {pagination.page} / {Math.max(1, pagination.totalPages)}</span><div>
              <button className="button button-ghost" disabled={loading || page <= 1} onClick={() => setPage((p) => p - 1)}>← Trước</button>
              <button className="button button-ghost" disabled={loading || page >= pagination.totalPages} onClick={() => setPage((p) => p + 1)}>Tiếp →</button>
            </div></div>
          </section>
          <footer className="footer">TaskFlow · Full-stack demo · React + Express.js + PostgreSQL</footer>
        </div>
      </main>
      {modalOpen && <TaskForm task={editingTask} saving={saving} error={formError} onSave={saveTask} onClose={() => setModalOpen(false)} />}
    </div>
  );
}
```


## `web/src/components/TaskCard.jsx`


```jsx
const priorities = {
  low: ['Thấp', 'low'],
  medium: ['Trung bình', 'medium'],
  high: ['Cao', 'high']
};

function displayDate(value) {
  if (!value) return null;
  return new Date(`${value}T12:00:00`).toLocaleDateString('vi-VN');
}

export default function TaskCard({ task, busy, onToggle, onEdit, onDelete }) {
  const priority = priorities[task.priority] || priorities.medium;
  const now = new Date();
  const localToday = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-${String(now.getDate()).padStart(2, '0')}`;
  const isOverdue = !task.completed && task.due_date && task.due_date < localToday;
  return (
    <article className={`task-card ${task.completed ? 'task-done' : ''}`}>
      <input className="task-check" type="checkbox" checked={task.completed}
        disabled={busy} aria-label={`Đánh dấu hoàn thành: ${task.title}`}
        onChange={() => onToggle(task)} />
      <div className="task-body">
        <div className="task-head">
          <h3>{task.title}</h3>
          <span className={`priority priority-${priority[1]}`}>{priority[0]}</span>
        </div>
        {task.description && <p className="task-description">{task.description}</p>}
        <div className="task-meta">
          <span>#{task.id}</span>
          {task.due_date && <span className={isOverdue ? 'overdue' : ''}>
            ◷ {isOverdue ? 'Quá hạn · ' : 'Hạn · '}{displayDate(task.due_date)}
          </span>}
          {task.completed && <span className="completed-label">✓ Đã hoàn thành</span>}
        </div>
      </div>
      <div className="task-actions">
        <button className="icon-button" aria-label={`Sửa ${task.title}`} title="Chỉnh sửa" onClick={() => onEdit(task)} disabled={busy}>✎</button>
        <button className="icon-button danger" aria-label={`Xóa ${task.title}`} title="Xóa" onClick={() => onDelete(task)} disabled={busy}>✕</button>
      </div>
    </article>
  );
}
```


## `web/src/components/TaskForm.jsx`


```jsx
import { useEffect, useState } from 'react';

const empty = { title: '', description: '', priority: 'medium', due_date: '', completed: false };

export default function TaskForm({ task, saving, error, onSave, onClose }) {
  const [form, setForm] = useState(empty);
  useEffect(() => {
    setForm(task ? {
      title: task.title,
      description: task.description,
      priority: task.priority,
      due_date: task.due_date || '',
      completed: task.completed
    } : empty);
  }, [task]);

  function update(key, value) {
    setForm((current) => ({ ...current, [key]: value }));
  }

  function submit(event) {
    event.preventDefault();
    onSave({
      title: form.title.trim(),
      description: form.description.trim(),
      priority: form.priority,
      due_date: form.due_date || null,
      completed: form.completed
    });
  }

  return (
    <div className="modal-backdrop" onMouseDown={(event) => {
      if (event.target === event.currentTarget && !saving) onClose();
    }}>
      <section className="modal" role="dialog" aria-modal="true" aria-labelledby="task-dialog-title">
        <div className="modal-heading">
          <div>
            <span className="eyebrow">TASK DETAILS</span>
            <h2 id="task-dialog-title">{task ? 'Chỉnh sửa công việc' : 'Thêm công việc mới'}</h2>
          </div>
          <button type="button" className="icon-button" aria-label="Đóng" onClick={onClose} disabled={saving}>✕</button>
        </div>
        <form onSubmit={submit} className="task-form">
          <label>Tên công việc <span aria-hidden="true">*</span>
            <input autoFocus required maxLength={120} value={form.title}
              onChange={(event) => update('title', event.target.value)} placeholder="Ví dụ: Hoàn thành báo cáo" />
          </label>
          <label>Mô tả
            <textarea rows={3} maxLength={2000} value={form.description}
              onChange={(event) => update('description', event.target.value)} placeholder="Ghi chú hoặc chi tiết công việc..." />
          </label>
          <div className="form-grid">
            <label>Mức ưu tiên
              <select value={form.priority} onChange={(event) => update('priority', event.target.value)}>
                <option value="low">Thấp</option>
                <option value="medium">Trung bình</option>
                <option value="high">Cao</option>
              </select>
            </label>
            <label>Hạn hoàn thành
              <input type="date" value={form.due_date}
                onChange={(event) => update('due_date', event.target.value)} />
            </label>
          </div>
          <label className="checkbox-label"><input type="checkbox" checked={form.completed}
            onChange={(event) => update('completed', event.target.checked)} /> Đã hoàn thành</label>
          {error && <div className="form-error" role="alert">{error}</div>}
          <div className="form-actions">
            <button type="button" className="button button-ghost" onClick={onClose} disabled={saving}>Hủy</button>
            <button className="button button-primary" type="submit" disabled={saving}>
              {saving ? 'Đang lưu…' : task ? 'Lưu thay đổi' : 'Tạo công việc'}
            </button>
          </div>
        </form>
      </section>
    </div>
  );
}
```


## `web/src/lib/api.js`


```javascript
const baseUrl = (import.meta.env.VITE_API_URL || '').replace(/\/$/, '');

async function request(path, options = {}) {
  let response;
  try {
    response = await fetch(`${baseUrl}${path}`, {
      ...options,
      headers: {
        Accept: 'application/json',
        ...(options.body ? { 'Content-Type': 'application/json' } : {}),
        ...options.headers
      }
    });
  } catch {
    throw new Error('Không thể kết nối API. Kiểm tra backend và địa chỉ VITE_API_URL.');
  }
  if (response.status === 204) return null;
  const data = await response.json().catch(() => null);
  if (!response.ok) {
    const details = data?.error?.details;
    const detail = Array.isArray(details) && details.length
      ? ` (${details[0].field}: ${details[0].message})` : '';
    throw new Error(`${data?.error?.message || `HTTP ${response.status}`}${detail}`);
  }
  return data;
}

export const taskApi = {
  list(filters) {
    const search = new URLSearchParams();
    for (const [key, value] of Object.entries(filters)) {
      if (value !== '' && value !== undefined) search.set(key, String(value));
    }
    return request(`/api/tasks?${search}`);
  },
  stats: () => request('/api/tasks/stats'),
  create: (task) => request('/api/tasks', { method: 'POST', body: JSON.stringify(task) }),
  update: (id, task) => request(`/api/tasks/${id}`, { method: 'PATCH', body: JSON.stringify(task) }),
  remove: (id) => request(`/api/tasks/${id}`, { method: 'DELETE' })
};
```


## `web/src/main.jsx`


```jsx
import React from 'react';
import { createRoot } from 'react-dom/client';
import App from './App.jsx';
import './styles.css';

createRoot(document.getElementById('root')).render(
  <React.StrictMode><App /></React.StrictMode>
);
```


## `web/src/styles.css`


```css
:root{font-family:Inter,ui-sans-serif,system-ui,-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif;color:#1b2540;background:#f6f8fc;font-synthesis:none;text-rendering:optimizeLegibility;-webkit-font-smoothing:antialiased}*{box-sizing:border-box}body{margin:0;min-width:320px}button,input,select,textarea{font:inherit}button{cursor:pointer}button:disabled{opacity:.55;cursor:not-allowed}button:focus-visible,input:focus-visible,textarea:focus-visible,select:focus-visible{outline:3px solid rgba(54,92,245,.25);outline-offset:2px}h1,h2,h3,p{margin-top:0}.app-shell{display:flex;min-height:100vh}.sidebar{width:255px;flex:none;background:#fff;border-right:1px solid #e9edf5;padding:32px 21px;display:flex;flex-direction:column;position:sticky;top:0;height:100vh}.brand{display:flex;align-items:center;gap:12px;font-weight:800;font-size:22px;letter-spacing:-.9px;padding:0 12px 51px}.brand-icon{width:38px;height:38px;display:grid;place-items:center;background:#365cf5;border-radius:11px;color:white;font-size:27px;box-shadow:0 8px 18px #365cf532}.brand-accent{color:#365cf5}.nav-caption,.eyebrow{font-size:10px;letter-spacing:2px;color:#8793aa;font-weight:800}.nav-caption{padding-left:14px;margin-bottom:16px}.nav-item{text-decoration:none;padding:13px 14px;color:#8793aa;font-weight:650;font-size:13px;border-radius:10px;margin-bottom:7px;display:flex;gap:13px;align-items:center}.nav-item span{font-size:20px}.nav-item.active{background:#edf2ff;color:#365cf5}.sidebar-bottom{margin-top:auto}.sidebar-tip{border-radius:15px;background:#eef2ff;padding:19px 18px;color:#30457e}.sidebar-tip strong{font-size:13px}.sidebar-tip p{font-size:12px;line-height:1.7;margin:9px 0 0;color:#66769d}.sidebar-version{font-size:10px;color:#9ba5b6;display:block;margin:24px 8px 0}.main-content{min-width:0;flex:1}.topbar{height:78px;background:rgba(255,255,255,.75);border-bottom:1px solid #e9edf5;display:flex;align-items:center;justify-content:flex-end;padding:0 50px}.topbar-right{font-size:12px;font-weight:650;display:flex;align-items:center;gap:12px;color:#7b879e}.live-dot{width:7px;height:7px;background:#38ba84;border-radius:100%}.avatar{height:34px;width:34px;border-radius:50%;display:grid;place-items:center;background:#e7ebff;color:#365cf5;font-size:11px;font-weight:800;margin-left:8px}.mobile-brand{display:none}.content-container{max-width:1190px;padding:44px 50px 24px;margin:auto}.page-heading{display:flex;justify-content:space-between;align-items:center;gap:20px;margin-bottom:32px}.page-heading h1{font-size:32px;letter-spacing:-1.4px;margin:11px 0 9px;font-weight:800}.heading-period{color:#365cf5}.page-heading p{font-size:13px;color:#8a96a9;margin-bottom:0}.button{border:1px solid transparent;border-radius:10px;display:inline-flex;align-items:center;justify-content:center;gap:8px;font-weight:700;font-size:12px;padding:12px 17px;white-space:nowrap;transition:background .15s,transform .15s}.button:active:not(:disabled){transform:translateY(1px)}.button-primary{background:#365cf5;color:#fff;box-shadow:0 6px 14px #365cf51e}.button-primary:hover{background:#2549d2}.button-ghost{border-color:#e7ecf5;background:white;color:#62708b}.button-ghost:hover{background:#f4f6fb}.add-button{height:43px;padding-inline:20px}.add-button span{font-size:21px;font-weight:400;line-height:0}.stats-grid{display:grid;grid-template-columns:repeat(4,minmax(0,1fr));gap:16px;margin-bottom:35px}.stat-card{background:#fff;border:1px solid #e9edf6;border-radius:15px;min-height:114px;display:flex;align-items:center;gap:15px;padding:22px 18px;box-shadow:0 5px 15px #1f355707}.stat-icon{width:47px;height:47px;flex:none;border-radius:13px;display:grid;place-items:center;font-size:24px;font-weight:600}.tone-blue{background:#e9eeff;color:#4261ed}.tone-amber{background:#fff5df;color:#e19a21}.tone-green{background:#dff7ed;color:#27a875}.tone-red{background:#ffebec;color:#dc5869}.stat-number{font-size:28px;line-height:1;font-weight:800;letter-spacing:-.8px;margin-bottom:8px}.stat-label{font-size:11px;color:#8c97aa;font-weight:600}.task-section{background:#fff;border:1px solid #e9edf5;box-shadow:0 7px 24px #1f355707;border-radius:18px;padding:25px 25px 20px}.section-heading{display:flex;align-items:start;justify-content:space-between;margin-bottom:21px}.section-heading h2{font-size:19px;letter-spacing:-.5px;margin-bottom:6px}.section-heading p{font-size:12px;color:#a2acbc;margin-bottom:0}.section-badge{background:#f0f3ff;color:#526bd1;border-radius:7px;font-size:11px;font-weight:700;padding:7px 11px}.filters{display:flex;gap:12px;margin-bottom:19px}.filters .search-field{flex:1;min-width:140px;position:relative}.search-field span{position:absolute;left:14px;top:50%;transform:translateY(-55%);font-size:24px;color:#a5afbf}.filters input,.filters select{height:43px;background:#fbfcff;border:1px solid #e7ebf3;border-radius:9px;color:#586781;font-size:12px;outline:0}.filters input{width:100%;padding:0 13px 0 43px}.filters select{padding:0 30px 0 12px}.filters input:focus,.filters select:focus{border-color:#88a0ff}.task-list{display:flex;flex-direction:column;gap:10px}.task-card{display:flex;align-items:flex-start;gap:14px;border:1px solid #ebeff5;border-radius:11px;padding:19px 17px;transition:background .15s,border-color .15s}.task-card:hover{background:#fcfdff;border-color:#dfe5f1}.task-check{appearance:none;-webkit-appearance:none;width:20px;height:20px;flex:none;border:2px solid #d0d9e7;border-radius:6px;margin:1px 2px 0 0;cursor:pointer;background:#fff}.task-check:checked{border-color:#39ba88;background:#39ba88;position:relative}.task-check:checked::after{content:'✓';position:absolute;inset:-4px 0 0 2px;color:#fff;font-size:18px;font-weight:700}.task-body{flex:1;min-width:0}.task-head{display:flex;align-items:center;gap:12px;flex-wrap:wrap}.task-head h3{font-size:13px;line-height:1.5;margin:0;font-weight:750;color:#263650;word-break:break-word}.task-done .task-head h3{text-decoration:line-through;color:#9fa8b6}.priority{display:inline-flex;padding:5px 9px;border-radius:6px;font-size:10px;font-weight:750;white-space:nowrap}.priority-low{background:#edf4ff;color:#5485bd}.priority-medium{background:#fff5df;color:#bf8b2a}.priority-high{background:#ffeded;color:#d45f5f}.task-description{font-size:12px;color:#8290a4;line-height:1.55;margin:5px 0 9px;overflow-wrap:anywhere}.task-meta{display:flex;gap:14px;flex-wrap:wrap;color:#a0aabc;font-size:10px;font-weight:600;margin-top:8px}.task-meta .overdue{color:#d45f5f}.task-meta .completed-label{color:#25a678}.task-actions{display:flex;gap:3px}.icon-button{width:32px;height:32px;background:transparent;border:0;border-radius:7px;color:#8b99ae;font-size:18px;display:grid;place-items:center}.icon-button:hover{background:#eff2fc;color:#365cf5}.icon-button.danger:hover{background:#fff0f2;color:#cf5661}.pagination{display:flex;justify-content:space-between;align-items:center;margin-top:19px;padding-top:16px;border-top:1px solid #edf0f5;color:#8d98aa;font-size:11px}.pagination>div{display:flex;gap:8px}.pagination .button{padding:9px 13px}.notice{padding:13px 16px;border-radius:10px;font-size:12px;margin-bottom:14px}.error-notice{background:#fff1f1;border:1px solid #ffcbcb;color:#a83a3a}.inline-retry{background:none;border:0;text-decoration:underline;color:inherit;font-weight:700}.empty-state{text-align:center;padding:60px 15px;color:#9aa5b6;font-size:13px}.empty-state h3{color:#344561;margin:0 0 8px}.empty-state p{margin:0 0 20px}.empty-icon{margin:auto auto 14px;font-size:34px;width:60px;height:60px;display:grid;place-items:center;border-radius:16px;background:#eff3ff;color:#365cf5}.footer{text-align:center;font-size:10px;color:#a5afbe;padding-top:28px}.modal-backdrop{position:fixed;inset:0;background:rgba(14,25,52,.42);z-index:20;display:flex;align-items:center;justify-content:center;padding:20px}.modal{background:white;border-radius:18px;box-shadow:0 25px 90px #0b183c3a;width:min(100%,520px);max-height:95vh;overflow-y:auto;padding:27px}.modal-heading{display:flex;align-items:start;justify-content:space-between;margin-bottom:23px}.modal-heading h2{font-size:22px;letter-spacing:-.5px;margin:7px 0 0}.task-form{display:flex;flex-direction:column;gap:18px}.task-form label{display:flex;flex-direction:column;gap:8px;color:#384760;font-size:12px;font-weight:700}.task-form label span{color:#df6073}.task-form input:not([type=checkbox]),.task-form textarea,.task-form select{width:100%;min-height:43px;border:1px solid #e1e6ef;border-radius:9px;padding:10px 12px;font-size:13px;background:#fcfdff;color:#334155}.task-form textarea{resize:vertical}.form-grid{display:grid;grid-template-columns:1fr 1fr;gap:14px}.task-form .checkbox-label{flex-direction:row;align-items:center;font-weight:600;gap:9px}.checkbox-label input{width:17px;height:17px;accent-color:#365cf5}.form-actions{display:flex;justify-content:flex-end;gap:10px;padding-top:10px}.form-error{padding:12px;background:#fff0f2;color:#aa3d4a;border:1px solid #ffd1d8;border-radius:8px;font-size:12px}@media(max-width:1100px){.sidebar{width:215px}.content-container{padding:34px 25px}.topbar{padding:0 25px}.stat-card{padding:18px 12px;gap:10px}.stat-icon{width:40px;height:40px}.stat-number{font-size:25px}}@media(max-width:850px){.sidebar{display:none}.mobile-brand{display:block;font-weight:800;font-size:19px;color:#365cf5}.topbar{justify-content:space-between;height:67px}.content-container{padding:27px 20px}.stats-grid{grid-template-columns:repeat(2,minmax(0,1fr))}}@media(max-width:600px){.topbar-right{font-size:0}.topbar-right .avatar{font-size:11px}.page-heading{align-items:flex-start;flex-direction:column}.page-heading h1{font-size:27px}.stats-grid{gap:9px}.stat-card{min-height:90px;padding:13px}.stat-icon{height:39px;width:39px}.stat-number{font-size:24px}.stat-label{font-size:10px}.task-section{padding:19px 13px}.filters{flex-wrap:wrap}.filters .search-field{flex-basis:100%}.filters select{flex:1;min-width:0}.task-card{padding:14px 11px;gap:8px}.task-actions{flex-direction:column}.form-grid{grid-template-columns:1fr}}
```


## `web/vite.config.js`


```javascript
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    strictPort: true,
    proxy: {
      '/api': 'http://localhost:4000',
      '/health': 'http://localhost:4000'
    }
  }
});
```
