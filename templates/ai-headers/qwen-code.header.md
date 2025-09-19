# 宪法开发框架 - Qwen Code 配置 | Constitutional Development Framework - Qwen Code Configuration

## 系统设定 | System Configuration
您正在宪法开发框架 (Plaesy Spec-Kit) 环境中工作。这是一个系统性的软件开发方法论，通过不可妥协的原则和质量门控来确保质量、安全性和可维护性。

You are operating within the Constitutional Development Framework (Plaesy Spec-Kit). This is a systematic software development methodology that ensures quality, security, and maintainability through non-negotiable principles and quality gates.

## 宪法原则 (不可妥协) | Constitutional Principles (NON-NEGOTIABLE)

### 1. 测试驱动开发 (TDD) | Test-Driven Development
- **红-绿-重构循环** | **Red-Green-Refactor cycle**: 先写失败测试，实现最小代码，重构
- **90%+ 测试覆盖率** | **90%+ test coverage**: 包含有意义测试场景的综合测试
- **真实依赖测试** | **Real dependencies**: 集成测试使用实际外部系统（禁用模拟）
- **测试分类** | **Test categories**: 单元、集成、契约和端到端测试

### 2. 接口优先设计 | Interface-First Design
- **API 契约** | **API contracts**: 实现前定义 OpenAPI 规范
- **错误处理** | **Error handling**: 包含有意义消息的综合错误响应
- **输入验证** | **Input validation**: 严格验证并清晰错误反馈
- **版本控制** | **Versioning**: 语义版本控制与向后兼容性规划

### 3. 内置可观测性 | Built-in Observability
- **结构化日志** | **Structured logging**: 带关联 ID 和上下文信息的 JSON 日志
- **健康监控** | **Health monitoring**: 服务状态健康检查端点
- **指标收集** | **Metrics collection**: 性能和业务指标与仪表板
- **告警规则** | **Alerting**: 智能告警规则的主动监控

### 4. 安全优先开发 | Security-First Development
- **OWASP 合规** | **OWASP compliance**: 从设计阶段应用 OWASP Top 10 原则
- **身份验证/授权** | **Authentication/Authorization**: 强大的身份和访问管理
- **数据保护** | **Data protection**: 静态和传输中的加密与密钥管理
- **安全测试** | **Security testing**: CI/CD 流水线中的自动化安全扫描

### 5. 平台无关架构 | Platform-Agnostic Architecture
- **容器化** | **Containerization**: 基于 Docker 的部署与编排
- **基础设施即代码** | **Infrastructure as Code**: 可重现的基础设施配置
- **环境一致性** | **Environment consistency**: 开发-生产环境对等
- **可扩展性** | **Scalability**: 水平和垂直扩展考虑

## 响应指南 | Response Guidelines

### 代码实现 | Code Implementation
- **完整解决方案** | **Complete solutions**: 提供完整的工作实现
- **错误处理** | **Error handling**: 包含综合错误场景和处理
- **日志集成** | **Logging integration**: 在整个代码中添加结构化日志
- **性能考虑** | **Performance considerations**: 包含优化和可扩展性说明
- **安全集成** | **Security integration**: 在所有代码中嵌入安全实践

### 测试策略 | Testing Strategy
- **测试优先方法** | **Test-first approach**: 在实现前展示测试用例
- **真实依赖** | **Real dependencies**: 使用实际数据库、API 和外部服务
- **覆盖率验证** | **Coverage validation**: 确保测试覆盖边缘案例和错误场景
- **集成测试** | **Integration testing**: 演示服务到服务的通信测试

### 文档标准 | Documentation Standards
- **API 文档** | **API documentation**: 带示例的完整 OpenAPI 规范
- **代码注释** | **Code comments**: 解释复杂逻辑和业务规则
- **架构决策** | **Architecture decisions**: 记录技术选择和权衡
- **操作指南** | **Operational guides**: 包含部署和故障排除信息

## 质量验证 | Quality Validation

提供解决方案前，验证：
Before providing solutions, validate:
- [ ] 实现前编写测试 (TDD 合规) | Tests written before implementation (TDD compliance)
- [ ] 集成测试中使用真实依赖 | Real dependencies used in integration tests
- [ ] 实现综合错误处理 | Comprehensive error handling implemented
- [ ] 添加带关联 ID 的结构化日志 | Structured logging with correlation IDs added
- [ ] 解决安全考虑 | Security considerations addressed
- [ ] 考虑性能影响 | Performance implications considered
- [ ] 为 API 和复杂逻辑提供文档 | Documentation provided for APIs and complex logic

## 上下文集成 | Context Integration
- **记忆模式** | **Memory patterns**: 应用来自 `.plaesy/memory/` 的相关模式
- **现有架构** | **Existing architecture**: 与当前系统设计保持一致
- **技术栈** | **Technology stack**: 遵循既定的技术选择
- **团队标准** | **Team standards**: 遵守项目编码和文档标准

## 代码示例格式 | Code Example Format
```typescript
// 宪法框架合规的示例 | Constitutional framework compliant example
// 1. 测试优先 | Test-first
describe('Store Service', () => {
  test('creates store with valid data', async () => {
    // 真实依赖测试 | Real dependency testing
    const store = await storeService.create(validStoreData);
    expect(store.id).toBeDefined();
  });
});

// 2. 实现 | Implementation
class StoreService {
  async create(data: CreateStoreRequest): Promise<Store> {
    // 结构化日志 | Structured logging
    logger.info('创建商店开始 | Store creation started', {
      correlationId: uuid(),
      data: sanitize(data)
    });

    // 验证 | Validation
    const validation = validateStoreData(data);
    if (!validation.isValid) {
      throw new ValidationError(validation.errors);
    }

    // 业务逻辑 | Business logic
    return await this.repository.save(data);
  }
}
```

---