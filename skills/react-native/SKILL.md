---
name: react-native
description: "React Native standards for mobile apps, with Spec-Kit constitutional compliance. Use when building or reviewing React Native code."
metadata:
  applies-to: "**/*.jsx, **/*.tsx, **/*.js, **/*.ts, **/*.json, package.json"
---

# React Native Development Instructions for Spec-Kit Framework

Apply `.plaesy/instructions/ui-ux-design-principles.md` (hierarchy, feedback,
affordance, Fitts's law — touch-target sizing matters even more on mobile) to
every screen you build — this file covers framework mechanics, not
interaction design.

## Constitutional Compliance for React Native

### Library-First Architecture

- Standalone packages per feature via npm workspaces
- Platform-specific functionality as separate native-module libraries
- Reusable component libraries with Storybook
- Clean public interfaces exported through index files

```typescript

// packages/user-management/src/index.ts
export { UserService } from './services/UserService';
export { UserRepository } from './repositories/UserRepository';
export { UserProfile } from './components/UserProfile';
export type { User, UserCreateRequest } from './types';

// CLI interface
export { default as cli } from './cli';
```

### CLI Interface Protocol for React Native

- **Node.js CLI**: Create CLI tools using Commander.js or Yargs
- **JSON Output**: Support structured output for automation
- **Cross-Platform**: Ensure CLI works on all development platforms

```typescript

// packages/user-management/src/cli.ts
#!/usr/bin/env node

import { Command } from 'commander';
import { UserService } from './services/UserService';

const program = new Command();

program
  .name('user-management')
  .description('User management CLI tool')
  .version('1.0.0');

program
  .command('create')
  .description('Create a new user')
  .option('--json', 'Output in JSON format')
  .argument('<email>', 'User email')
  .argument('<name>', 'User name')
  .action(async (email: string, name: string, options) => {
    try {
      const userService = new UserService();
      const user = await userService.createUser({ email, name });

      if (options.json) {
        console.log(JSON.stringify({ user, status: 'success' }));
      } else {
        console.log(`User created: ${user.name} (${user.email})`);
      }
      process.exit(0);
    } catch (error) {
      if (options.json) {
        console.log(JSON.stringify({ error: error.message, status: 'error' }));
      } else {
        console.error(`Error: ${error.message}`);
      }
      process.exit(1);
    }
  });

program.parse();
```

### Test-First Development (TDD)

- **Jest Configuration**: Configure Jest for React Native testing
- **Testing Library**: Use React Native Testing Library for component tests
- **Detox**: Use Detox for end-to-end testing on real devices
- **Real APIs**: Test against real backend services

```typescript

// packages/user-management/__tests__/UserService.test.ts
import { UserService } from '../src/services/UserService';
import { ApiClient } from '../src/api/ApiClient';

describe('UserService Integration Tests', () => {
  let userService: UserService;

  beforeEach(() => {
    // Use real API client, not mocked
    const apiClient = new ApiClient({
      baseURL: process.env.TEST_API_URL || 'https://api.example.com',
      timeout: 5000,
    });
    userService = new UserService(apiClient);
  });

  describe('createUser', () => {
    it('should create user with valid data', async () => {
      // Arrange
      const userData = {
        email: 'test@example.com',
        name: 'Test User',
      };

      // Act
      const user = await userService.createUser(userData);

      // Assert
      expect(user.email).toBe(userData.email);
      expect(user.name).toBe(userData.name);
      expect(user.id).toBeDefined();
    });

    it('should throw error for invalid email', async () => {
      // This test should FAIL first, then implement validation
      const userData = {
        email: 'invalid-email',
        name: 'Test User',
      };

      await expect(userService.createUser(userData))
        .rejects
        .toThrow('Invalid email format');
    });
  });
});
```

### Component Testing with Real Dependencies

```typescript

// packages/user-management/__tests__/UserProfile.test.tsx
import React from 'react';
import { render, waitFor, screen } from '@testing-library/react-native';
import { UserProfile } from '../src/components/UserProfile';
import { UserService } from '../src/services/UserService';

describe('UserProfile Component', () => {
  let userService: UserService;

  beforeEach(() => {
    // Use real service, not mocked
    userService = new UserService();
  });

  it('should display user information when loaded', async () => {
    // Arrange
    const userId = 1;

    // Act
    render(<UserProfile userId={userId} userService={userService} />);

    // Assert
    await waitFor(() => {
      expect(screen.getByText(/John Doe/)).toBeTruthy();
      expect(screen.getByText(/john@example.com/)).toBeTruthy();
    });
  });

  it('should show loading state initially', () => {
    render(<UserProfile userId={1} userService={userService} />);

    expect(screen.getByTestId('loading-indicator')).toBeTruthy();
  });
});
```

## Constitutional Compliance Checklist

### For Every React Native Package/Feature

- [ ] **Library Structure**: Package has clean public API exports
- [ ] **CLI Interface**: Node.js CLI tool with --json support
- [ ] **TDD Compliance**: Tests written first with failing tests committed
- [ ] **Real Dependencies**: Integration tests use real APIs/services
- [ ] **Observability**: Structured logging and error tracking
- [ ] **Performance**: Optimized for mobile performance
- [ ] **Security**: Secure storage and network practices
- [ ] **Platform Support**: Works on both iOS and Android

### Code Review Checklist

- [ ] Component tests use React Native Testing Library
- [ ] E2E tests use Detox with real app builds
- [ ] No hardcoded API URLs or secrets
- [ ] Error boundaries implemented for critical components
- [ ] Performance optimizations applied (memo, callbacks)
- [ ] Platform-specific code properly handled
- [ ] Constitutional compliance verified

---
Ensures React Native development within Spec-Kit stays constitutionally compliant while following React Native best practices and mobile conventions.

## Reference

Read `references/best-practices.md` for React Native best practices and the project structure guide when implementing or reorganizing app code.
