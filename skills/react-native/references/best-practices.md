# React Native Best Practices and Project Structure

## React Native Best Practices

### State Management

- Redux Toolkit for complex state
- React Query for server state
- Context API for simple local state
- Zustand as a lightweight alternative

```typescript

// packages/user-management/src/store/userSlice.ts
import { createSlice, createAsyncThunk } from '@reduxjs/toolkit';
import { UserService } from '../services/UserService';

export const fetchUser = createAsyncThunk(
  'user/fetchUser',
  async (userId: number, { rejectWithValue }) => {
    try {
      const userService = new UserService();
      return await userService.getUser(userId);
    } catch (error) {
      return rejectWithValue(error.message);
    }
  }
);

const userSlice = createSlice({
  name: 'user',
  initialState: {
    data: null,
    loading: false,
    error: null,
  },
  reducers: {
    clearUser: (state) => {
      state.data = null;
      state.error = null;
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(fetchUser.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(fetchUser.fulfilled, (state, action) => {
        state.loading = false;
        state.data = action.payload;
      })
      .addCase(fetchUser.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload as string;
      });
  },
});

export const { clearUser } = userSlice.actions;
export default userSlice.reducer;
```

### Performance Optimization

- **FlatList Optimization**: Use getItemLayout and keyExtractor
- **Image Optimization**: Use FastImage for better image performance
- **Bundle Splitting**: Use code splitting for large features
- **Memory Management**: Avoid memory leaks in components

```typescript

// packages/user-management/src/components/UserList.tsx
import React, { memo, useCallback } from 'react';
import { FlatList, View, Text } from 'react-native';
import FastImage from 'react-native-fast-image';

interface User {
  id: number;
  name: string;
  email: string;
  avatar: string;
}

interface Props {
  users: User[];
  onUserPress: (user: User) => void;
}

const UserListItem = memo(({ user, onPress }: { user: User; onPress: (user: User) => void }) => {
  const handlePress = useCallback(() => onPress(user), [user, onPress]);

  return (
    <View style={styles.item}>
      <FastImage
        source={{ uri: user.avatar }}
        style={styles.avatar}
        resizeMode={FastImage.resizeMode.cover}
      />
      <View style={styles.info}>
        <Text style={styles.name}>{user.name}</Text>
        <Text style={styles.email}>{user.email}</Text>
      </View>
    </View>
  );
});

export const UserList = memo(({ users, onUserPress }: Props) => {
  const renderItem = useCallback(
    ({ item }: { item: User }) => (
      <UserListItem user={item} onPress={onUserPress} />
    ),
    [onUserPress]
  );

  const keyExtractor = useCallback((item: User) => item.id.toString(), []);

  const getItemLayout = useCallback(
    (data: any, index: number) => ({
      length: 80, // Item height
      offset: 80 * index,
      index,
    }),
    []
  );

  return (
    <FlatList
      data={users}
      renderItem={renderItem}
      keyExtractor={keyExtractor}
      getItemLayout={getItemLayout}
      removeClippedSubviews={true}
      maxToRenderPerBatch={10}
      windowSize={10}
    />
  );
});
```

### Error Handling & Observability

- **Structured Logging**: Use Flipper or React Native Logs for development
- **Crash Reporting**: Integrate Crashlytics or Bugsnag
- **Performance Monitoring**: Use Flipper or custom performance tracking

```typescript

// packages/core/src/utils/Logger.ts
interface LogEntry {
  level: 'info' | 'warn' | 'error' | 'debug';
  message: string;
  timestamp: string;
  extra?: Record<string, any>;
}

class Logger {
  private static instance: Logger;

  static getInstance(): Logger {
    if (!Logger.instance) {
      Logger.instance = new Logger();
    }
    return Logger.instance;
  }

  private log(level: LogEntry['level'], message: string, extra?: Record<string, any>) {
    const entry: LogEntry = {
      level,
      message,
      timestamp: new Date().toISOString(),
      extra,
    };

    // Log to console in development
    if (__DEV__) {
      console.log(JSON.stringify(entry, null, 2));
    }

    // Send to crash reporting service
    if (level === 'error') {
      // crashlytics().recordError(new Error(message));
    }

    // Send to analytics
    // analytics().logEvent('app_log', entry);
  }

  info(message: string, extra?: Record<string, any>) {
    this.log('info', message, extra);
  }

  warn(message: string, extra?: Record<string, any>) {
    this.log('warn', message, extra);
  }

  error(message: string, extra?: Record<string, any>) {
    this.log('error', message, extra);
  }

  debug(message: string, extra?: Record<string, any>) {
    this.log('debug', message, extra);
  }
}

export const logger = Logger.getInstance();
```

### Platform-Specific Code

- **Platform Detection**: Use Platform.OS for conditional logic
- **Native Modules**: Create platform-specific implementations
- **File Extensions**: Use .ios.ts and .android.ts for platform-specific files

```typescript

// packages/platform/src/services/PlatformService.ts
import { Platform, NativeModules } from 'react-native';

class PlatformService {
  async getDeviceInfo(): Promise<any> {
    if (Platform.OS === 'ios') {
      return await NativeModules.IOSDeviceInfo.getDeviceInfo();
    } else if (Platform.OS === 'android') {
      return await NativeModules.AndroidDeviceInfo.getDeviceInfo();
    }
    throw new Error('Unsupported platform');
  }

  getStoragePath(): string {
    return Platform.select({
      ios: NativeModules.IOSFileManager.documentsPath,
      android: NativeModules.AndroidFileManager.externalStoragePath,
      default: '/tmp',
    });
  }
}

export const platformService = new PlatformService();
```

### Security Best Practices

- **Secure Storage**: Use Keychain (iOS) and Keystore (Android)
- **Network Security**: Implement certificate pinning
- **Code Obfuscation**: Use Hermes and enable ProGuard

```typescript

// packages/security/src/services/SecureStorage.ts
import { Platform } from 'react-native';
import * as Keychain from 'react-native-keychain';

class SecureStorage {
  async setItem(key: string, value: string): Promise<void> {
    try {
      await Keychain.setSecureValue({
        key,
        value,
        securityLevel: Platform.OS === 'ios'
          ? Keychain.SECURITY_LEVEL.SECURE_SOFTWARE
          : Keychain.SECURITY_LEVEL.SECURE_HARDWARE,
      });
    } catch (error) {
      throw new Error(`Failed to store secure item: ${error.message}`);
    }
  }

  async getItem(key: string): Promise<string | null> {
    try {
      const result = await Keychain.getSecureValue(key);
      return result || null;
    } catch (error) {
      if (error.code === 'item_not_found') {
        return null;
      }
      throw new Error(`Failed to retrieve secure item: ${error.message}`);
    }
  }

  async removeItem(key: string): Promise<void> {
    try {
      await Keychain.deleteSecureValue(key);
    } catch (error) {
      throw new Error(`Failed to remove secure item: ${error.message}`);
    }
  }
}

export const secureStorage = new SecureStorage();
```

## Project Structure

### Recommended Monorepo Structure

```text

react-native-app/
├── apps/
│   └── mobile/              # Main React Native app
│       ├── android/
│       ├── ios/
│       ├── src/
│       └── package.json
├── packages/
│   ├── core/               # Core utilities and types
│   ├── ui-components/      # Shared UI components
│   ├── user-management/    # Feature package
│   ├── authentication/    # Feature package
│   └── api-client/        # API integration
├── tools/
│   ├── metro-config/      # Metro bundler configuration
│   └── eslint-config/     # Shared ESLint configuration
└── package.json           # Root package.json
```

### Package.json Configuration

```json

{
  "name": "react-native-spec-kit-app",
  "version": "1.0.0",
  "private": true,
  "workspaces": [
    "apps/*",
    "packages/*"
  ],
  "scripts": {
    "start": "cd apps/mobile && npm start",
    "android": "cd apps/mobile && npm run android",
    "ios": "cd apps/mobile && npm run ios",
    "test": "jest",
    "test:e2e": "detox test",
    "lint": "eslint .",
    "type-check": "tsc --noEmit",
    "build:ios": "cd apps/mobile && npm run build:ios",
    "build:android": "cd apps/mobile && npm run build:android"
  },
  "devDependencies": {
    "@react-native/metro-config": "^0.72.0",
    "@types/jest": "^29.5.0",
    "@types/react": "^18.2.0",
    "@types/react-native": "^0.72.0",
    "detox": "^20.0.0",
    "jest": "^29.5.0",
    "typescript": "^5.0.0"
  }
}
```
