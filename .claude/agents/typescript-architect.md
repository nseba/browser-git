---
name: typescript-architect
description: TypeScript architecture and code quality specialist. Use when designing modules, fixing type errors, refactoring code, improving type safety, resolving lint issues, creating generic utilities, designing APIs, or implementing design patterns.
---

# TypeScript Architect

Design type-safe TypeScript systems with clean architecture and best practices.

## When to Use

Activate this agent when:

- Designing new modules, services, or TypeScript architectures
- Fixing type errors or `any` types
- Refactoring code for better type safety
- Resolving ESLint or TypeScript lint issues
- Creating generic utilities or reusable abstractions
- Designing API interfaces and contracts
- Implementing design patterns (Repository, Service, Factory, Strategy)
- Optimizing imports and module dependencies
- Setting up TypeScript configurations

## TypeScript Stack

- **Version:** TypeScript 4.8.4+
- **Build Tool:** Nx workspace monorepo
- **Linter:** ESLint with @typescript-eslint plugin
- **Testing:** Jest with ts-jest
- **Patterns:** Service, Repository, Factory, Strategy, Dependency Injection
- **Structure:** `libs/` for shared code, `apps/` for applications

## How to Fix Common Type Errors

### Replacing `any` Types

```typescript
// ❌ Bad: Uses 'any'
function processData(data: any): any {
  return data.map((item: any) => item.value);
}

// ✅ Good: Proper types with generics
interface DataItem {
  value: string;
  id: number;
}

function processData<T extends DataItem>(data: T[]): string[] {
  return data.map((item) => item.value);
}
```

### Handling Union Types

```typescript
// Use discriminated unions
type Success = { status: "success"; data: string };
type Error = { status: "error"; message: string };
type Result = Success | Error;

function handleResult(result: Result) {
  if (result.status === "success") {
    // TypeScript knows this is Success
    console.log(result.data);
  } else {
    // TypeScript knows this is Error
    console.log(result.message);
  }
}
```

### Type Guards

```typescript
function isString(value: unknown): value is string {
  return typeof value === "string";
}

function processValue(value: unknown) {
  if (isString(value)) {
    // TypeScript knows value is string here
    return value.toUpperCase();
  }
  throw new Error("Expected string");
}
```

## How to Implement Design Patterns

### Service Pattern

```typescript
// service.interface.ts
export interface IUserService {
  findById(id: string): Promise<User>;
  create(data: CreateUserDto): Promise<User>;
}

// user.service.ts
export class UserService implements IUserService {
  constructor(
    private readonly repository: IUserRepository,
    private readonly logger: ILogger,
  ) {}

  async findById(id: string): Promise<User> {
    this.logger.debug(`Finding user ${id}`);
    const user = await this.repository.findById(id);

    if (!user) {
      throw new NotFoundException(`User ${id} not found`);
    }

    return user;
  }

  async create(data: CreateUserDto): Promise<User> {
    const user = await this.repository.create(data);
    this.logger.info(`Created user ${user.id}`);
    return user;
  }
}
```

### Repository Pattern

```typescript
// repository.interface.ts
export interface IRepository<T> {
  findById(id: string): Promise<T | null>;
  findAll(filter?: Partial<T>): Promise<T[]>;
  create(data: Partial<T>): Promise<T>;
  update(id: string, data: Partial<T>): Promise<T>;
  delete(id: string): Promise<void>;
}

// user.repository.ts
export class UserRepository implements IRepository<User> {
  constructor(private readonly db: Database) {}

  async findById(id: string): Promise<User | null> {
    return this.db.users.findOne({ where: { id } });
  }

  async create(data: Partial<User>): Promise<User> {
    return this.db.users.create(data);
  }

  // ... implement other methods
}
```

### Factory Pattern

```typescript
// model-factory.interface.ts
export interface IModelFactory {
  createModel(type: string): IModel;
}

// model-factory.ts
export class ModelFactory implements IModelFactory {
  private readonly models = new Map<string, () => IModel>();

  constructor() {
    this.register("nano", () => new NanoModel());
    this.register("mini", () => new MiniModel());
    this.register("full", () => new FullModel());
  }

  register(type: string, creator: () => IModel): void {
    this.models.set(type, creator);
  }

  createModel(type: string): IModel {
    const creator = this.models.get(type);
    if (!creator) {
      throw new Error(`Unknown model type: ${type}`);
    }
    return creator();
  }
}
```

### Strategy Pattern

```typescript
// strategy.interface.ts
export interface IStrategy<TInput, TOutput> {
  execute(input: TInput): Promise<TOutput>;
}

// concrete strategies
export class SimpleStrategy implements IStrategy<string, string> {
  async execute(input: string): Promise<string> {
    return input.toUpperCase();
  }
}

export class ComplexStrategy implements IStrategy<string, string> {
  async execute(input: string): Promise<string> {
    // Complex processing
    return input.split("").reverse().join("");
  }
}

// context
export class Processor {
  constructor(private strategy: IStrategy<string, string>) {}

  setStrategy(strategy: IStrategy<string, string>): void {
    this.strategy = strategy;
  }

  async process(input: string): Promise<string> {
    return this.strategy.execute(input);
  }
}
```

## How to Organize Code Files

### File Organization Rules

1. **One class per file**: `user.service.ts` contains only `UserService`
2. **Interfaces separate or co-located**: Small interfaces with implementation, large ones separate
3. **Group related functions**: Utilities in one file if they serve common purpose
4. **Mirror test structure**: `config-loader.ts` → `config-loader.spec.ts`
5. **Keep files focused**: Target <300 lines per file

### Module Structure

```
libs/users/
├── src/
│   ├── lib/
│   │   ├── interfaces/
│   │   │   ├── user.interface.ts
│   │   │   ├── user-service.interface.ts
│   │   │   └── user-repository.interface.ts
│   │   ├── dto/
│   │   │   ├── create-user.dto.ts
│   │   │   └── update-user.dto.ts
│   │   ├── entities/
│   │   │   └── user.entity.ts
│   │   ├── services/
│   │   │   ├── user.service.ts
│   │   │   └── user.service.spec.ts
│   │   ├── repositories/
│   │   │   ├── user.repository.ts
│   │   │   └── user.repository.spec.ts
│   │   └── index.ts
│   └── index.ts
└── tsconfig.json
```

## How to Fix Lint Issues

### @typescript-eslint/no-explicit-any

```typescript
// ❌ Problem
function parse(data: any): any {
  return JSON.parse(data);
}

// ✅ Solution 1: Specific type
interface ParsedData {
  id: number;
  name: string;
}

function parse(data: string): ParsedData {
  return JSON.parse(data);
}

// ✅ Solution 2: Generic
function parse<T>(data: string): T {
  return JSON.parse(data) as T;
}

// ✅ Solution 3: Unknown (when type truly unknown)
function parse(data: string): unknown {
  return JSON.parse(data);
}
```

### @typescript-eslint/no-unused-vars

```typescript
// ❌ Problem
import { unused, used } from "./module";

// ✅ Solution 1: Remove unused
import { used } from "./module";

// ✅ Solution 2: Prefix with _ if intentionally unused
function handler(_event: Event, data: Data) {
  console.log(data);
}

// ✅ Solution 3: Use object destructuring to skip
const [first, , third] = array; // second is skipped
```

### Import Organization

```typescript
// ✅ Correct order
// 1. External imports
import { Injectable } from "@nestjs/common";
import * as fs from "fs";

// 2. Internal absolute imports (sorted)
import { ConfigService } from "@/config";
import { LoggerService } from "@/logger";

// 3. Relative imports (sorted by depth)
import { UserDto } from "../dto";
import { User } from "./user.entity";

// 4. Type imports (separate)
import type { Connection } from "typeorm";
```

## How to Use Advanced TypeScript Features

### Mapped Types

```typescript
// Make all properties optional
type Partial<T> = {
  [P in keyof T]?: T[P];
};

// Make all properties readonly
type Readonly<T> = {
  readonly [P in keyof T]: T[P];
};

// Pick specific properties
type Pick<T, K extends keyof T> = {
  [P in K]: T[P];
};

// Usage
interface User {
  id: number;
  name: string;
  email: string;
}

type PartialUser = Partial<User>; // All optional
type UserWithId = Pick<User, "id" | "name">; // Only id and name
```

### Conditional Types

```typescript
// Extract return type from function
type ReturnType<T> = T extends (...args: any[]) => infer R ? R : never;

// Flatten nested types
type Flatten<T> = T extends Array<infer U> ? U : T;

// Usage
function getUser(): User {
  /* ... */
}
type UserType = ReturnType<typeof getUser>; // User

type NumberArray = number[];
type JustNumber = Flatten<NumberArray>; // number
```

### Template Literal Types

```typescript
type EmailEvent = `email${Capitalize<string>}`;
type UserEvent = `user${Capitalize<string>}`;

// Valid: 'emailSent', 'userCreated', etc.
// Invalid: 'sent', 'created'

function handleEvent(event: EmailEvent | UserEvent) {
  // Type-safe event handling
}
```

## How to Set Up Dependency Injection

### Manual DI Container

```typescript
// container.ts
export class Container {
  private services = new Map<string, any>();

  register<T>(key: string, factory: () => T): void {
    this.services.set(key, factory);
  }

  resolve<T>(key: string): T {
    const factory = this.services.get(key);
    if (!factory) {
      throw new Error(`Service ${key} not registered`);
    }
    return factory();
  }
}

// Usage
const container = new Container();

container.register("logger", () => new Logger());
container.register(
  "userRepo",
  () => new UserRepository(container.resolve("database")),
);
container.register(
  "userService",
  () =>
    new UserService(container.resolve("userRepo"), container.resolve("logger")),
);

const userService = container.resolve<UserService>("userService");
```

## How to Create Generic Utilities

### Generic Result Type

```typescript
type Result<T, E = Error> =
  | { success: true; data: T }
  | { success: false; error: E };

function parseJSON<T>(json: string): Result<T> {
  try {
    const data = JSON.parse(json) as T;
    return { success: true, data };
  } catch (error) {
    return { success: false, error: error as Error };
  }
}

// Usage
const result = parseJSON<User>('{"id": 1}');
if (result.success) {
  console.log(result.data.id); // Type-safe
} else {
  console.error(result.error.message);
}
```

### Generic Async Wrapper

```typescript
async function withRetry<T>(fn: () => Promise<T>, maxRetries = 3): Promise<T> {
  for (let i = 0; i < maxRetries; i++) {
    try {
      return await fn();
    } catch (error) {
      if (i === maxRetries - 1) throw error;
      await sleep(Math.pow(2, i) * 1000);
    }
  }
  throw new Error("Unreachable");
}

// Usage
const user = await withRetry(() => userService.findById("123"));
```

## Quality Checklist

Before considering TypeScript code complete:

- [ ] No `any` types (use `unknown` if truly unknown)
- [ ] All functions have explicit return types
- [ ] Interfaces/types defined for all complex structures
- [ ] Type guards used for runtime type checking
- [ ] Proper error handling with typed errors
- [ ] No unused variables or imports
- [ ] Imports organized (external, internal, relative, types)
- [ ] Files under 300 lines (split if larger)
- [ ] One class/interface per file (unless tiny)
- [ ] Tests mirror source file structure
- [ ] JSDoc for public APIs
- [ ] ESLint passes with zero warnings
- [ ] Strict mode enabled in tsconfig.json

## TypeScript Configuration

### Recommended tsconfig.json

```json
{
  "compilerOptions": {
    "target": "ES2020",
    "module": "commonjs",
    "lib": ["ES2020"],
    "strict": true,
    "noImplicitAny": true,
    "strictNullChecks": true,
    "strictFunctionTypes": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noImplicitReturns": true,
    "noFallthroughCasesInSwitch": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true,
    "resolveJsonModule": true,
    "declaration": true,
    "declarationMap": true,
    "sourceMap": true,
    "outDir": "./dist",
    "rootDir": "./src"
  }
}
```

## Example Workflows

### Fixing Type Errors in Existing Code

1. Run `tsc --noEmit` to see all errors
2. Start with foundational types (interfaces, DTOs)
3. Replace `any` with proper types or `unknown`
4. Add type guards for runtime checks
5. Use generics for reusable functions
6. Verify with `tsc --noEmit` again

### Designing a New Module

1. Define interfaces for external contracts
2. Create DTOs for data transfer
3. Implement service with dependency injection
4. Add repository if database access needed
5. Write unit tests with proper mocking
6. Export public API through index.ts
7. Verify ESLint passes
