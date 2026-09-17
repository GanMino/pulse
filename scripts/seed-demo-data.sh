#!/bin/bash
# ==========================================
# Pulse Mock 数据生成器
# ==========================================
# 直接往 SQLite 插入示例场景 + 历史测试运行
# 启动 Pulse 后立刻看到真实感的数据
# ==========================================

set -e

DB_PATH="${HOME}/.pulse/data/pulse.db"

if [ ! -f "$DB_PATH" ]; then
    echo "❌ 数据库不存在,请先运行 Pulse 一次以初始化"
    exit 1
fi

# 检查 sqlite3 命令
if ! command -v sqlite3 &> /dev/null; then
    echo "❌ sqlite3 命令未找到"
    echo "安装:brew install sqlite (macOS)"
    exit 1
fi

echo "→ 准备向 $DB_PATH 插入演示数据..."

# 默认用户
sqlite3 "$DB_PATH" <<'SQL'
-- 插入默认用户
INSERT OR IGNORE INTO users (uuid, username, email, display_name, role, created_at, updated_at)
VALUES ('local-user-uuid', 'local', 'local@pulse.dev', 'Local User', 'admin', datetime('now'), datetime('now'));

-- 插入默认项目
INSERT OR IGNORE INTO projects (uuid, name, description, color, owner_id, created_at, updated_at)
SELECT 'default-project-uuid', 'Default', '默认项目', '#3370FF', id, datetime('now'), datetime('now')
FROM users WHERE username = 'local';

-- 清理已有演示数据(避免重复)
DELETE FROM scenarios WHERE source = 'demo';
DELETE FROM test_runs WHERE triggered_by IN (SELECT id FROM users WHERE username = 'demo');

-- 演示场景 1:登录流程
INSERT INTO scenarios (
    uuid, project_id, name, description, config, source, status, version, created_by, created_at, updated_at
) VALUES (
    'demo-scenario-1',
    (SELECT id FROM projects WHERE name = 'Default'),
    '⭐ 登录流程压测',
    '模拟电商登录接口,验证 1000 并发下的响应能力',
    '{
        "load": {"vus": 100, "duration": "5m", "targetRPS": 0, "rampUp": {"type": "linear", "duration": "30s"}},
        "requests": [
            {"name": "登录", "method": "POST", "url": "/api/login", "headers": {"Content-Type": "application/json"}, "body": {"username": "{{user}}", "password": "{{pass}}"}, "extractors": {"token": "$.data.token"}, "assertions": {"status": [200], "maxLatencyMs": 200}},
            {"name": "获取用户", "method": "GET", "url": "/api/user/profile", "headers": {"Authorization": "Bearer {{token}}"}, "assertions": {"status": [200]}}
        ],
        "variables": {"user": ["alice", "bob", "charlie"], "pass": "123456"},
        "thresholds": {"p95Ms": 200, "errorRate": 0.01, "minRPS": 500}
    }',
    'demo',
    'active',
    3,
    (SELECT id FROM users WHERE username = 'local'),
    datetime('now', '-7 days'),
    datetime('now', '-1 day')
);

-- 演示场景 2:商品列表
INSERT INTO scenarios (
    uuid, project_id, name, description, config, source, status, version, created_by, created_at, updated_at
) VALUES (
    'demo-scenario-2',
    (SELECT id FROM projects WHERE name = 'Default'),
    '商品列表查询',
    'GET /api/products 压测,模拟用户浏览商品',
    '{
        "load": {"vus": 50, "duration": "2m"},
        "requests": [
            {"name": "查询商品", "method": "GET", "url": "/api/products?page=1&size=20", "assertions": {"status": [200], "maxLatencyMs": 100}}
        ],
        "thresholds": {"p95Ms": 100, "errorRate": 0.005}
    }',
    'demo',
    'active',
    1,
    (SELECT id FROM users WHERE username = 'local'),
    datetime('now', '-5 days'),
    datetime('now', '-5 days')
);

-- 演示场景 3:大促容量测试
INSERT INTO scenarios (
    uuid, project_id, name, description, config, source, status, version, created_by, created_at, updated_at
) VALUES (
    'demo-scenario-3',
    (SELECT id FROM projects WHERE name = 'Default'),
    '大促容量测试',
    '双 11 大促前容量评估,500 VUs 模拟流量峰值',
    '{
        "load": {"vus": 500, "duration": "10m", "rampUp": {"type": "step", "duration": "1m"}},
        "requests": [
            {"name": "首页", "method": "GET", "url": "/api/home", "assertions": {"status": [200]}},
            {"name": "搜索", "method": "GET", "url": "/api/search?q=phone"},
            {"name": "商品详情", "method": "GET", "url": "/api/products/123"},
            {"name": "加购物车", "method": "POST", "url": "/api/cart", "body": {"productId": 1, "qty": 1}},
            {"name": "下单", "method": "POST", "url": "/api/order", "assertions": {"status": [200, 201]}}
        ],
        "thresholds": {"p95Ms": 500, "errorRate": 0.02, "minRPS": 1000}
    }',
    'demo',
    'active',
    5,
    (SELECT id FROM users WHERE username = 'local'),
    datetime('now', '-30 days'),
    datetime('now', '-15 days')
);

-- 演示场景 4:草稿
INSERT INTO scenarios (
    uuid, project_id, name, description, config, source, status, version, created_by, created_at, updated_at
) VALUES (
    'demo-scenario-4',
    (SELECT id FROM projects WHERE name = 'Default'),
    '支付链路',
    '草稿:正在规划中',
    '{
        "load": {"vus": 10, "duration": "1m"},
        "requests": [
            {"name": "支付", "method": "POST", "url": "/api/pay", "body": {"orderId": 1, "amount": 99.9}}
        ]
    }',
    'demo',
    'draft',
    1,
    (SELECT id FROM users WHERE username = 'local'),
    datetime('now', '-2 days'),
    datetime('now', '-2 days')
);

-- 演示场景 5:HAR 导入示例
INSERT INTO scenarios (
    uuid, project_id, name, description, config, source, status, version, created_by, created_at, updated_at
) VALUES (
    'demo-scenario-5',
    (SELECT id FROM projects WHERE name = 'Default'),
    '从 Chrome HAR 录制导入',
    '由 Chrome DevTools 录制后通过 HAR 导入自动生成',
    '{
        "load": {"vus": 20, "duration": "3m"},
        "requests": [
            {"name": "GET /api/feed", "method": "GET", "url": "/api/feed", "headers": {"Authorization": "Bearer ..."}},
            {"name": "POST /api/like", "method": "POST", "url": "/api/like", "body": {"postId": 1}},
            {"name": "GET /api/profile/me", "method": "GET", "url": "/api/profile/me"}
        ]
    }',
    'har_import',
    'active',
    2,
    (SELECT id FROM users WHERE username = 'local'),
    datetime('now', '-10 days'),
    datetime('now', '-8 days')
);
SQL

echo "✓ 插入了 5 个演示场景"

# 插入历史测试运行
echo "→ 生成历史测试运行记录..."

sqlite3 "$DB_PATH" <<'SQL'
-- 运行 1:登录流程 - 7 天前,成功
INSERT INTO test_runs (
    uuid, scenario_id, project_id, status, started_at, finished_at,
    duration_ms, runtime_config, summary, trigger_type, triggered_by, created_at
) VALUES (
    'run-demo-1',
    (SELECT id FROM scenarios WHERE uuid = 'demo-scenario-1'),
    (SELECT id FROM projects WHERE name = 'Default'),
    'completed',
    datetime('now', '-7 days', '-5 hours'),
    datetime('now', '-7 days'),
    1800000,
    '{"vus": 100, "duration": "30m"}',
    '{
        "TotalRequests": 287341,
        "TotalErrors": 1247,
        "ErrorRate": 0.0043,
        "AvgRPS": 159.6,
        "PeakRPS": 234.5,
        "P50Ms": 65.3,
        "P90Ms": 121.5,
        "P95Ms": 142.8,
        "P99Ms": 231.2,
        "MaxMs": 512.3,
        "AvgMs": 73.5,
        "StatusCodes": {"200": 286094, "201": 0, "500": 1247}
    }',
    'manual',
    (SELECT id FROM users WHERE username = 'local'),
    datetime('now', '-7 days')
);

-- 运行 2:登录流程 - 3 天前,失败(超时)
INSERT INTO test_runs (
    uuid, scenario_id, project_id, status, started_at, finished_at,
    duration_ms, runtime_config, summary, trigger_type, triggered_by, created_at
) VALUES (
    'run-demo-2',
    (SELECT id FROM scenarios WHERE uuid = 'demo-scenario-1'),
    (SELECT id FROM projects WHERE name = 'Default'),
    'failed',
    datetime('now', '-3 days', '-3 hours'),
    datetime('now', '-3 days'),
    600000,
    '{"vus": 200, "duration": "10m"}',
    '{
        "TotalRequests": 142000,
        "TotalErrors": 12800,
        "ErrorRate": 0.09,
        "AvgRPS": 236.7,
        "P50Ms": 89.2,
        "P95Ms": 523.1,
        "P99Ms": 891.4,
        "MaxMs": 2304.7,
        "StatusCodes": {"200": 129200, "500": 12800}
    }',
    'manual',
    (SELECT id FROM users WHERE username = 'local'),
    datetime('now', '-3 days')
);

-- 运行 3:商品列表 - 5 天前,成功
INSERT INTO test_runs (
    uuid, scenario_id, project_id, status, started_at, finished_at,
    duration_ms, summary, trigger_type, triggered_by, created_at
) VALUES (
    'run-demo-3',
    (SELECT id FROM scenarios WHERE uuid = 'demo-scenario-2'),
    (SELECT id FROM projects WHERE name = 'Default'),
    'completed',
    datetime('now', '-5 days', '-8 hours'),
    datetime('now', '-5 days'),
    120000,
    '{
        "TotalRequests": 58421,
        "TotalErrors": 12,
        "ErrorRate": 0.0002,
        "AvgRPS": 486.8,
        "P50Ms": 28.5,
        "P90Ms": 62.3,
        "P95Ms": 78.1,
        "P99Ms": 134.7,
        "MaxMs": 287.4,
        "StatusCodes": {"200": 58409}
    }',
    'manual',
    (SELECT id FROM users WHERE username = 'local'),
    datetime('now', '-5 days')
);

-- 运行 4:大促容量 - 15 天前,失败
INSERT INTO test_runs (
    uuid, scenario_id, project_id, status, started_at, finished_at,
    duration_ms, summary, trigger_type, triggered_by, created_at
) VALUES (
    'run-demo-4',
    (SELECT id FROM scenarios WHERE uuid = 'demo-scenario-3'),
    (SELECT id FROM projects WHERE name = 'Default'),
    'failed',
    datetime('now', '-15 days', '-2 hours'),
    datetime('now', '-15 days'),
    900000,
    '{
        "TotalRequests": 198000,
        "TotalErrors": 19800,
        "ErrorRate": 0.10,
        "AvgRPS": 220,
        "P95Ms": 1245.7,
        "P99Ms": 2341.2,
        "MaxMs": 5891.4,
        "StatusCodes": {"200": 178200, "500": 19800}
    }',
    'manual',
    (SELECT id FROM users WHERE username = 'local'),
    datetime('now', '-15 days')
);

-- 运行 5:登录流程 - 1 天前,完美通过
INSERT INTO test_runs (
    uuid, scenario_id, project_id, status, started_at, finished_at,
    duration_ms, summary, trigger_type, triggered_by, created_at
) VALUES (
    'run-demo-5',
    (SELECT id FROM scenarios WHERE uuid = 'demo-scenario-1'),
    (SELECT id FROM projects WHERE name = 'Default'),
    'completed',
    datetime('now', '-1 day', '-5 hours'),
    datetime('now', '-1 day'),
    1800000,
    '{
        "TotalRequests": 312450,
        "TotalErrors": 234,
        "ErrorRate": 0.0007,
        "AvgRPS": 173.6,
        "P50Ms": 58.2,
        "P90Ms": 112.8,
        "P95Ms": 138.4,
        "P99Ms": 215.6,
        "MaxMs": 423.1,
        "StatusCodes": {"200": 312216}
    }',
    'manual',
    (SELECT id FROM users WHERE username = 'local'),
    datetime('now', '-1 day')
);

-- 运行 6:HAR 导入场景 - 2 天前
INSERT INTO test_runs (
    uuid, scenario_id, project_id, status, started_at, finished_at,
    duration_ms, summary, trigger_type, triggered_by, created_at
) VALUES (
    'run-demo-6',
    (SELECT id FROM scenarios WHERE uuid = 'demo-scenario-5'),
    (SELECT id FROM projects WHERE name = 'Default'),
    'completed',
    datetime('now', '-2 days', '-10 hours'),
    datetime('now', '-2 days'),
    180000,
    '{
        "TotalRequests": 21600,
        "TotalErrors": 45,
        "ErrorRate": 0.002,
        "AvgRPS": 120,
        "P50Ms": 42.1,
        "P95Ms": 95.3,
        "P99Ms": 156.7,
        "MaxMs": 312.5,
        "StatusCodes": {"200": 21555}
    }',
    'manual',
    (SELECT id FROM users WHERE username = 'local'),
    datetime('now', '-2 days')
);

-- 运行 7:登录流程 - 12 小时前
INSERT INTO test_runs (
    uuid, scenario_id, project_id, status, started_at, finished_at,
    duration_ms, summary, trigger_type, triggered_by, created_at
) VALUES (
    'run-demo-7',
    (SELECT id FROM scenarios WHERE uuid = 'demo-scenario-1'),
    (SELECT id FROM projects WHERE name = 'Default'),
    'completed',
    datetime('now', '-12 hours', '-3 hours'),
    datetime('now', '-12 hours'),
    600000,
    '{
        "TotalRequests": 102341,
        "TotalErrors": 87,
        "ErrorRate": 0.0008,
        "AvgRPS": 170.6,
        "P50Ms": 62.4,
        "P95Ms": 145.7,
        "P99Ms": 224.8,
        "MaxMs": 487.2,
        "StatusCodes": {"200": 102254}
    }',
    'manual',
    (SELECT id FROM users WHERE username = 'local'),
    datetime('now', '-12 hours')
);

-- 创建对应的 reports
INSERT INTO reports (uuid, test_run_id, summary, status, created_at)
SELECT
    'report-' || uuid,
    id,
    summary,
    'ready',
    created_at
FROM test_runs WHERE uuid LIKE 'run-demo-%';
SQL

echo "✓ 插入了 7 条历史测试运行"
echo ""
echo "═══════════════════════════════════════"
echo "  ✅ 演示数据已生成!"
echo "═══════════════════════════════════════"
echo ""
echo "现在启动 Pulse 即可看到:"
echo "  • 5 个示例场景(登录 / 商品 / 大促 / 草稿 / HAR导入)"
echo "  • 7 条历史测试运行记录(包含成功 / 失败 / 不同时间)"
echo "  • 丰富的统计数据(总请求 100万+,不同 RPS / 延迟)"
echo ""
echo "下一步:"
echo "  wails dev        # 开发模式"
echo "  # 或"
echo "  wails build && open build/bin/pulse.app"
echo ""