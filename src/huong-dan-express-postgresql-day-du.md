# Hướng dẫn Express.js + PostgreSQL theo Controller – Service – Repository

Bản hướng dẫn này chứa **toàn bộ 21 file của project**, kèm hướng dẫn chạy và giải thích kiến trúc. Nó được tạo đồng bộ với file ZIP đi kèm.

## 1. Công nghệ và kiến trúc

- Node.js 22 trở lên, Express.js 5, `pg`, PostgreSQL 17 trong Docker, Zod 4, Helmet, CORS, dotenv, Supertest.
- Luồng request: `Client → Route → Validation → Controller → Service → Repository → PostgreSQL`, kết quả quay ngược lại.
- Controller đọc HTTP và trả status/JSON; Service thực thi quy tắc nghiệp vụ; Repository viết SQL.
- Tham số SQL dùng `$1, $2` thay vì nối chuỗi; không đưa input không được kiểm tra trực tiếp vào SQL.

## 2. Cách chạy dự án

Yêu cầu cài Node.js 22+, Docker và Docker Compose (hoặc một PostgreSQL server có sẵn).

```bash
npm install
cp .env.example .env
cp .env.test.example .env.test
docker compose up -d
npm run db:migrate
npm run db:migrate:test
npm run dev
```

Trên Windows PowerShell, thay `cp` bằng `Copy-Item`. Nếu cổng 5432 đã được dùng thì chỉnh `compose.yaml` và biến `DATABASE_URL` cho khớp. Nếu Docker volume đã có từ trước, init script tạo database test sẽ không chạy lại; tạo DB test riêng một cách thủ công.

- `GET http://localhost:3000/health`: kiểm tra HTTP server.
- `GET http://localhost:3000/ready`: kiểm tra database.
- `GET /api/tasks`: danh sách + phân trang; query hỗ trợ `completed=true|false`, `page`, `limit`.
- `GET /api/tasks/:id`: chi tiết.
- `POST /api/tasks`: tạo mới.
- `PATCH /api/tasks/:id`: cập nhật một phần.
- `DELETE /api/tasks/:id`: xóa.

Ví dụ tạo task:

```bash
curl -i -X POST http://localhost:3000/api/tasks \
  -H 'Content-Type: application/json' \
  -d '{"title":"Học Repository Pattern","description":"Lưu task vào PostgreSQL"}'
```

Sau khi chạy migration trên DB test:

```bash
npm test
```

Lưu ý: `.env.test` cần trỏ đến database độc lập có tên kết thúc bằng `_test`. Tests tạo/sửa/xóa dữ liệu.

## 3. Toàn bộ source code theo từng file

### 3.1. `package.json` – Node scripts và dependencies

````json
{
  "name": "express-postgresql-clean-architecture",
  "version": "1.0.0",
  "private": true,
  "type": "module",
  "engines": {
    "node": ">=22"
  },
  "scripts": {
    "dev": "node --watch src/server.js",
    "start": "node src/server.js",
    "db:migrate": "node scripts/migrate.js",
    "db:migrate:test": "node --env-file=.env.test scripts/migrate.js",
    "test": "node --env-file=.env.test --test"
  },
  "dependencies": {
    "cors": "^2.8.6",
    "dotenv": "^17.4.2",
    "express": "^5.3.0",
    "helmet": "^8.3.0",
    "pg": "^8.23.1",
    "zod": "^4.6.7"
  },
  "devDependencies": {
    "supertest": "^7.3.2"
  }
}
````

### 3.2. `.gitignore` – Không commit secrets và dependencies

````gitignore
node_modules/
.env
.env.test
.DS_Store
coverage/
````

### 3.3. `.env.example` – Cấu hình môi trường phát triển

````dotenv
NODE_ENV=development
PORT=3000
DATABASE_URL=postgresql://tasks_user:tasks_password_local@localhost:5432/tasks_db
CORS_ORIGIN=http://localhost:5173
````

### 3.4. `.env.test.example` – Database riêng cho integration tests

````dotenv
NODE_ENV=test
DATABASE_URL=postgresql://tasks_user:tasks_password_local@localhost:5432/tasks_db_test
CORS_ORIGIN=http://localhost:5173
````

### 3.5. `compose.yaml` – Tạo PostgreSQL qua Docker Compose

````yaml
services:
  postgres:
    image: postgres:17-alpine
    environment:
      POSTGRES_USER: tasks_user
      POSTGRES_PASSWORD: tasks_password_local
      POSTGRES_DB: tasks_db
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./db/init:/docker-entrypoint-initdb.d:ro
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U tasks_user -d tasks_db"]
      interval: 5s
      timeout: 3s
      retries: 10

volumes:
  postgres_data:
````

### 3.6. `db/init/001_create_test_database.sql` – Khởi tạo database test khi tạo Docker volume

````sql
-- Chạy một lần khi Docker tạo database volume mới.
CREATE DATABASE tasks_db_test;
````

### 3.7. `db/migrations/001_create_tasks.sql` – Schema bảng tasks

````sql
CREATE TABLE IF NOT EXISTS tasks (
  id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  title VARCHAR(120) NOT NULL CHECK (length(btrim(title)) > 0),
  description TEXT NOT NULL DEFAULT '' CHECK (char_length(description) <= 2000),
  completed BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tasks_created_at
  ON tasks (created_at DESC, id DESC);
````

### 3.8. `src/config/db.js` – Tạo PostgreSQL Pool

````javascript
import 'dotenv/config';
import pg from 'pg';

const { Pool } = pg;

if (!process.env.DATABASE_URL) {
  throw new Error('Missing DATABASE_URL environment variable');
}

// Tránh chạy test/migration test vào database phát triển hoặc production.
if (process.env.NODE_ENV === 'test') {
  const databaseName = decodeURIComponent(
    new URL(process.env.DATABASE_URL).pathname.slice(1)
  );
  if (!databaseName.endsWith('_test')) {
    throw new Error('Test database name must end with _test');
  }
}

export const pool = new Pool({
  connectionString: process.env.DATABASE_URL,
  max: 10,
  connectionTimeoutMillis: 5000,
  idleTimeoutMillis: 30000
});

// Bắt lỗi bất ngờ từ client đang rảnh trong pool.
pool.on('error', (error) => {
  console.error('Unexpected PostgreSQL pool error:', error);
});
````

### 3.9. `src/errors/app-error.js` – Định nghĩa lỗi nghiệp vụ

````javascript
export class AppError extends Error {
  constructor(statusCode, message, details = undefined) {
    super(message);
    this.name = 'AppError';
    this.statusCode = statusCode;
    this.details = details;
  }
}
````

### 3.10. `src/schemas/task.schema.js` – Schema validation với Zod

````javascript
import { z } from 'zod';

const positiveDecimal = z.string()
  .regex(/^[1-9]\d*$/, 'Must be a positive integer')
  .transform(Number)
  .pipe(z.int().positive());

export const taskParamsSchema = z.strictObject({
  id: positiveDecimal.pipe(z.number().max(2147483647))
});

export const listTasksQuerySchema = z.strictObject({
  completed: z.enum(['true', 'false'])
    .transform((value) => value === 'true')
    .optional(),
  page: positiveDecimal.default(1),
  limit: positiveDecimal.pipe(z.number().max(100)).default(10)
});

export const createTaskSchema = z.strictObject({
  title: z.string().trim().min(1).max(120),
  description: z.string().trim().max(2000).default(''),
  completed: z.boolean().default(false)
});

export const updateTaskSchema = z.strictObject({
  title: z.string().trim().min(1).max(120).optional(),
  description: z.string().trim().max(2000).optional(),
  completed: z.boolean().optional()
}).refine((data) => Object.keys(data).length > 0, {
  message: 'Provide at least one field to update'
});
````

### 3.11. `src/middleware/validate.js` – Middleware parse và validate

````javascript
import { AppError } from '../errors/app-error.js';

export function validate(schema, source = 'body') {
  return (req, res, next) => {
    const result = schema.safeParse(req[source]);

    if (!result.success) {
      const details = result.error.issues.map((issue) => ({
        field: issue.path.join('.') || source,
        message: issue.message
      }));

      throw new AppError(400, 'Invalid request data', details);
    }

    // Lưu dữ liệu đã parse vào res.locals, không ghi đè req.query.
    res.locals.input ??= {};
    res.locals.input[source] = result.data;
    next();
  };
}
````

### 3.12. `src/middleware/error-handler.js` – Middleware xử lý lỗi

````javascript
import { AppError } from '../errors/app-error.js';

export function notFound(req, res, next) {
  next(new AppError(404, 'Endpoint not found'));
}

export function errorHandler(err, req, res, next) {
  if (res.headersSent) {
    return next(err);
  }

  // Lỗi JSON parser từ express.json().
  if (err.type === 'entity.parse.failed') {
    return res.status(400).json({
      success: false,
      error: { message: 'Invalid JSON body' }
    });
  }

  if (err.type === 'entity.too.large') {
    return res.status(413).json({
      success: false,
      error: { message: 'Request body too large' }
    });
  }

  const knownError = err instanceof AppError;
  const statusCode = knownError ? err.statusCode : 500;

  if (!knownError) {
    console.error('Unexpected server error:', err);
  }

  res.status(statusCode).json({
    success: false,
    error: {
      message: knownError ? err.message : 'Internal server error',
      ...(knownError && err.details ? { details: err.details } : {})
    }
  });
}
````

### 3.13. `src/repositories/task.repository.js` – Tầng Repository - toàn bộ SQL

````javascript
import { pool } from '../config/db.js';

const TASK_COLUMNS =
  'id, title, description, completed, created_at, updated_at';

export const taskRepository = {
  async findMany({ completed, page, limit }) {
    const where = completed === undefined
      ? ''
      : 'WHERE completed = $1';
    const filterValues = completed === undefined ? [] : [completed];

    const values = [...filterValues, limit, (page - 1) * limit];
    const limitPosition = values.length - 1;
    const offsetPosition = values.length;

    const listSql = `
      SELECT ${TASK_COLUMNS}
      FROM tasks
      ${where}
      ORDER BY created_at DESC, id DESC
      LIMIT $${limitPosition}
      OFFSET $${offsetPosition}
    `;

    const countSql = `SELECT COUNT(*)::int AS total FROM tasks ${where}`;

    const [itemsResult, countResult] = await Promise.all([
      pool.query(listSql, values),
      pool.query(countSql, filterValues)
    ]);

    return {
      items: itemsResult.rows,
      total: countResult.rows[0].total
    };
  },

  async findById(id) {
    const result = await pool.query(
      `SELECT ${TASK_COLUMNS} FROM tasks WHERE id = $1`,
      [id]
    );
    return result.rows[0] ?? null;
  },

  async create({ title, description, completed }) {
    const result = await pool.query(
      `INSERT INTO tasks (title, description, completed)
       VALUES ($1, $2, $3)
       RETURNING ${TASK_COLUMNS}`,
      [title, description, completed]
    );
    return result.rows[0];
  },

  async update(id, changes) {
    const setters = [];
    const values = [];

    // Danh sách cột cố định, không đưa tên cột do client gửi vào SQL.
    for (const column of ['title', 'description', 'completed']) {
      if (Object.hasOwn(changes, column)) {
        values.push(changes[column]);
        setters.push(`${column} = $${values.length}`);
      }
    }

    setters.push('updated_at = NOW()');
    values.push(id);

    const result = await pool.query(
      `UPDATE tasks
       SET ${setters.join(', ')}
       WHERE id = $${values.length}
       RETURNING ${TASK_COLUMNS}`,
      values
    );

    return result.rows[0] ?? null;
  },

  async remove(id) {
    const result = await pool.query(
      'DELETE FROM tasks WHERE id = $1 RETURNING id',
      [id]
    );
    return result.rowCount > 0;
  }
};
````

### 3.14. `src/services/task.service.js` – Tầng Service - business logic

````javascript
import { taskRepository } from '../repositories/task.repository.js';
import { AppError } from '../errors/app-error.js';

export const taskService = {
  async list(query) {
    return taskRepository.findMany(query);
  },

  async getById(id) {
    const task = await taskRepository.findById(id);
    if (!task) {
      throw new AppError(404, 'Task not found');
    }
    return task;
  },

  async create(data) {
    return taskRepository.create(data);
  },

  async update(id, changes) {
    const task = await taskRepository.update(id, changes);
    if (!task) {
      throw new AppError(404, 'Task not found');
    }
    return task;
  },

  async remove(id) {
    const deleted = await taskRepository.remove(id);
    if (!deleted) {
      throw new AppError(404, 'Task not found');
    }
  }
};
````

### 3.15. `src/controllers/task.controller.js` – Tầng Controller - HTTP

````javascript
import { taskService } from '../services/task.service.js';

export const taskController = {
  async list(req, res) {
    const query = res.locals.input.query;
    const { items, total } = await taskService.list(query);

    res.json({
      success: true,
      data: items,
      pagination: {
        page: query.page,
        limit: query.limit,
        total,
        totalPages: Math.ceil(total / query.limit)
      }
    });
  },

  async getById(req, res) {
    const task = await taskService.getById(res.locals.input.params.id);
    res.json({ success: true, data: task });
  },

  async create(req, res) {
    const task = await taskService.create(res.locals.input.body);
    res.location(`/api/tasks/${task.id}`)
      .status(201)
      .json({ success: true, data: task });
  },

  async update(req, res) {
    const task = await taskService.update(
      res.locals.input.params.id,
      res.locals.input.body
    );
    res.json({ success: true, data: task });
  },

  async remove(req, res) {
    await taskService.remove(res.locals.input.params.id);
    res.status(204).end();
  }
};
````

### 3.16. `src/routes/task.routes.js` – Router đăng ký endpoint

````javascript
import { Router } from 'express';
import { taskController } from '../controllers/task.controller.js';
import { validate } from '../middleware/validate.js';
import {
  taskParamsSchema,
  listTasksQuerySchema,
  createTaskSchema,
  updateTaskSchema
} from '../schemas/task.schema.js';

const router = Router();

router.get('/', validate(listTasksQuerySchema, 'query'), taskController.list);
router.get('/:id', validate(taskParamsSchema, 'params'), taskController.getById);
router.post('/', validate(createTaskSchema), taskController.create);
router.patch(
  '/:id',
  validate(taskParamsSchema, 'params'),
  validate(updateTaskSchema),
  taskController.update
);
router.delete('/:id', validate(taskParamsSchema, 'params'), taskController.remove);

export default router;
````

### 3.17. `src/app.js` – Khởi tạo ứng dụng Express

````javascript
import 'dotenv/config';
import express from 'express';
import cors from 'cors';
import helmet from 'helmet';

import { pool } from './config/db.js';
import taskRoutes from './routes/task.routes.js';
import { errorHandler, notFound } from './middleware/error-handler.js';

const app = express();

app.disable('x-powered-by');
app.use(helmet());
app.use(cors({ origin: process.env.CORS_ORIGIN || 'http://localhost:5173' }));
app.use(express.json({ limit: '100kb' }));

// Liveness: HTTP server đang hoạt động.
app.get('/health', (req, res) => {
  res.json({ status: 'ok' });
});

// Readiness: có thể thực hiện truy vấn vào PostgreSQL.
app.get('/ready', async (req, res) => {
  await pool.query('SELECT 1');
  res.json({ status: 'ready' });
});

app.use('/api/tasks', taskRoutes);
app.use(notFound);
app.use(errorHandler);

export default app;
````

### 3.18. `src/server.js` – Bật server và shutdown an toàn

````javascript
import app from './app.js';
import { pool } from './config/db.js';

const PORT = Number(process.env.PORT || 3000);

async function start() {
  if (!Number.isInteger(PORT) || PORT < 1 || PORT > 65535) {
    throw new Error('PORT must be an integer between 1 and 65535');
  }

  // Kiểm tra database sẵn sàng trước khi mở HTTP server.
  await pool.query('SELECT 1');

  const server = app.listen(PORT, () => {
    console.log(`API running at http://localhost:${PORT}`);
  });

  function shutdown(signal) {
    console.log(`Received ${signal}; shutting down`);
    server.close(async (error) => {
      if (error) {
        console.error('HTTP server shutdown error:', error);
        process.exitCode = 1;
      }
      try {
        await pool.end();
      } catch (dbError) {
        console.error('Database shutdown error:', dbError);
        process.exitCode = 1;
      }
    });
  }

  process.once('SIGINT', () => shutdown('SIGINT'));
  process.once('SIGTERM', () => shutdown('SIGTERM'));
}

start().catch(async (error) => {
  console.error('Failed to start API:', error);
  process.exitCode = 1;
  await pool.end();
});
````

### 3.19. `scripts/migrate.js` – Chạy migration

````javascript
import { readFile } from 'node:fs/promises';
import { pool } from '../src/config/db.js';

try {
  const path = new URL('../db/migrations/001_create_tasks.sql', import.meta.url);
  const sql = await readFile(path, 'utf8');
  await pool.query(sql);
  console.log('Database migration completed.');
} catch (error) {
  console.error('Migration failed:', error);
  process.exitCode = 1;
} finally {
  await pool.end();
}
````

### 3.20. `tests/tasks.test.js` – Integration test qua Supertest

````javascript
import { test, after } from 'node:test';
import assert from 'node:assert/strict';
import request from 'supertest';
import app from '../src/app.js';
import { pool } from '../src/config/db.js';

after(async () => {
  await pool.end();
});

test('Create, read, update, filter and delete a task', async () => {
  const created = await request(app)
    .post('/api/tasks')
    .send({ title: 'Integration test', description: 'From Supertest' })
    .expect(201);

  const id = created.body.data.id;
  assert.equal(created.body.data.completed, false);

  try {
    const fetched = await request(app)
      .get(`/api/tasks/${id}`)
      .expect(200);
    assert.equal(fetched.body.data.title, 'Integration test');

    const updated = await request(app)
      .patch(`/api/tasks/${id}`)
      .send({ completed: true })
      .expect(200);
    assert.equal(updated.body.data.completed, true);

    const listed = await request(app)
      .get('/api/tasks?completed=true&page=1&limit=10')
      .expect(200);
    assert.ok(listed.body.data.some((task) => task.id === id));
    assert.equal(listed.body.pagination.page, 1);

    await request(app).delete(`/api/tasks/${id}`).expect(204);
    await request(app).get(`/api/tasks/${id}`).expect(404);
  } finally {
    // Dọn dữ liệu nếu một assertion phía trên thất bại.
    await pool.query('DELETE FROM tasks WHERE id = $1', [id]);
  }
});

test('Reject invalid payload and invalid ID', async () => {
  await request(app)
    .post('/api/tasks')
    .send({ title: '' })
    .expect(400);

  await request(app)
    .patch('/api/tasks/1')
    .send({})
    .expect(400);

  await request(app).get('/api/tasks/not-a-number').expect(400);
  await request(app).get('/api/tasks?limit=999').expect(400);
});

test('Respond to liveness and readiness checks', async () => {
  await request(app).get('/health').expect(200);
  await request(app).get('/ready').expect(200);
});
````

### 3.21. `README.md` – Hướng dẫn trong project

````markdown
# Express.js + PostgreSQL: Controller / Service / Repository

Yêu cầu: Node.js >= 22, Docker Compose (hoặc PostgreSQL có sẵn).

## Khởi chạy nhanh

```bash
npm install
cp .env.example .env
cp .env.test.example .env.test
# Nếu dùng Docker:
docker compose up -d
npm run db:migrate
npm run db:migrate:test
npm run dev
```

API: http://localhost:3000

- GET /health
- GET /ready
- GET /api/tasks?completed=false&page=1&limit=10
- GET /api/tasks/:id
- POST /api/tasks
- PATCH /api/tasks/:id
- DELETE /api/tasks/:id

Ví dụ:

```bash
curl -X POST http://localhost:3000/api/tasks \
  -H 'Content-Type: application/json' \
  -d '{"title":"Learn PostgreSQL","description":"Use repositories"}'
```

## Kiểm thử

```bash
npm test
```

**Cảnh báo:** `.env.test` phải chỉ tới database có tên kết thúc bằng `_test`. Test thực hiện INSERT/UPDATE/DELETE. Không chạy test trên database chứa dữ liệu thật.

Script migration 001 có thể chạy lại nhờ `IF NOT EXISTS`. Trong production nên sử dụng hệ thống quản lý lịch sử migration chuyên dụng.

Docker chỉ khởi tạo database test ở lần đầu tạo volume. Nếu đang tái sử dụng volume cũ, bạn phải tạo database test thủ công hoặc sử dụng volume mới. Không xóa Docker volume khi chứa dữ liệu cần giữ.

## Cấu trúc

- `routes`: URL + method + validation middleware
- `controllers`: HTTP request/response
- `services`: business rules
- `repositories`: SQL queries
- `schemas`: Zod schemas
- `middleware`: validation / errors
- `config/db.js`: PostgreSQL pool
- `scripts/migrate.js`: chạy SQL migration
- `tests`: integration tests qua Supertest + PostgreSQL riêng
````

## 4. Giải thích request từ đầu đến cuối

Ví dụ `PATCH /api/tasks/1` với JSON `{"completed":true}`:

1. Express nhận request và `express.json()` parse body.
2. Router khớp endpoint, kiểm tra `params` và `body` bằng Zod.
3. Controller lấy `res.locals.input.params.id` và `res.locals.input.body`, gọi Service.
4. Service gọi Repository cập nhật và trả 404 nếu task không tồn tại.
5. Repository sinh câu lệnh `UPDATE tasks SET completed = $1, updated_at = NOW() WHERE id = $2 RETURNING ...`.
6. PostgreSQL thực hiện SQL, Repository trả row mới; Controller gửi JSON và HTTP 200.
7. Nếu bất kỳ `await` nào throw/reject, Express 5 chuyển lỗi tới error handler.

## 5. Các điểm kỹ thuật cần chú ý

- `src/app.js` không gọi `listen()` để Supertest có thể import và gửi request trực tiếp.
- `src/server.js` kiểm tra kết nối DB trước khi bắt đầu lắng nghe và gọi `pool.end()` khi shutdown.
- Schema Zod dùng `z.strictObject()` nhằm từ chối trường không nằm trong API contract.
- List endpoint kiểm tra `page`/`limit` để tránh truy vấn không kiểm soát; `COUNT(*)::int` phù hợp quy mô tutorial.
- CORS không thay thế xác thực/phân quyền. Đây là demo Task CRUD chưa có authentication.
- Migration 001 dùng `IF NOT EXISTS` để chạy lại mà không lỗi. Dự án lớn nên chuyển sang migration tool có version tracking.
- SQL `UPDATE` chỉ dùng tên cột từ danh sách cố định do ứng dụng định nghĩa; giá trị đều là parameters.
- Docker password chỉ là mật khẩu demo local; không dùng trên production.
- Tất cả code đã qua `node --check`, nhưng không thể chạy E2E ở môi trường tạo file do thiếu PostgreSQL và quyền tải packages từ npm.

## 6. Các bước nâng cấp

- Tách migration có version tracking; thêm rollback khi cần.
- Thêm User module và session/JWT authentication, ownerId, authorization.
- Thêm integration tests xác nhận authorization, đồng thời kiểm thử toàn bộ validation branches.
- Thêm logging, rate limiting, request IDs, monitoring.
- Dùng transaction với `pool.connect()` và `BEGIN/COMMIT/ROLLBACK` khi một nghiệp vụ có nhiều SQL statements phải thành công hoặc thất bại cùng nhau.

## Tài liệu chính thức

- https://expressjs.com/en/5x/guide/routing/
- https://expressjs.com/en/5x/guide/error-handling/
- https://node-postgres.com/features/queries
- https://node-postgres.com/features/pooling
- https://zod.dev/api
