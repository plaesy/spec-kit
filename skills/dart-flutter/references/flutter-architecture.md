# Flutter Architecture Recommendations

Architecture best practices - treat as recommendations, not rules, and adapt to your app's requirements. Priority reflects how strongly the Flutter team recommends it:

- **Strongly recommend**: always implement for new apps; strongly consider refactoring existing apps unless it fundamentally clashes with the current approach
- **Recommend**: will likely improve your app
- **Conditional**: can improve your app in certain circumstances

## Separation of concerns

Separate your app into UI and data layers; within those, separate logic into classes by responsibility.

### Use a three-layer architecture (Domain, Data, Presentation)

**Strongly recommend** — the most important architectural principle:

#### Domain Layer

**Pure Dart only** — NO `import 'package:flutter'`.

- Entities: core domain models
- Use Cases: business logic orchestration
- Repository interfaces: abstract contracts

#### Data Layer

- Repositories: concrete implementations of domain contracts
- Data Sources: API clients, local DB drivers
- DTOs/Models: JSON serialization
- Caching, error handling, retry logic

#### Presentation Layer

- Widgets/Screens: UI components
- State management providers (Riverpod/BLoC)
- UI logic: navigation, theme, lifecycle

**Critical Rule**: Dependencies point INWARD only → Data → Domain ← Presentation (never reverse, never Data → Domain).

#### Use clearly defined data and UI layers

**Strongly recommend** — data layer exposes app data and holds most business logic. UI layer displays data and listens for user events, with separate classes for UI logic and widgets.

#### Use the repository pattern in the data layer

**Strongly recommend** — isolates data access logic from the rest of the app - an abstraction layer between business logic and storage (databases, APIs, file systems).
In practice: Repository classes and Service classes.

#### Use ViewModels and Views in the UI layer. (MVVM)

**Strongly recommend** — keeps widgets "dumb", making code much less error prone.

#### Use `ChangeNotifiers` and `Listenables` to handle widget updates

**Conditional** — state-management choice ultimately comes down to personal preference.

`ChangeNotifier` is part of the Flutter SDK - a convenient way for widgets to observe ViewModel changes.

#### Do not put logic in widgets

**Strongly recommend** — logic belongs in ViewModel methods. A view should only contain:

- Simple if-statements to show/hide widgets based on a ViewModel flag/nullable field
- Animation logic the widget must calculate
- Layout logic based on device info (screen size, orientation)
- Simple routing logic

#### Use a domain layer

**Conditional** — use in apps with complex logic requirements.

Only needed when logic complexity crowds ViewModels or logic repeats across them. Useful in very large apps; unnecessary overhead in most.

### Handling data

Handle data with care for code that's easier to understand, less error prone, and free of malformed/unexpected data.

#### Use unidirectional data flow

**Strongly recommend** — data updates flow only from data layer to UI layer; UI interactions are sent to the data layer for processing.

#### Use `Commands` to handle events from user interaction

**Recommend** — prevents rendering errors and standardizes how the UI layer sends events to the data layer.

#### Use immutable data models

**Strongly recommend** — ensures changes occur only in the proper place (usually data/domain layer) - immutable objects require a new instance to reflect change, preventing accidental UI-layer updates
and supporting unidirectional data flow.

#### Use freezed or built_value to generate immutable data models

**Recommend** — generate common model methods (JSON ser/des, deep equality, copy). Can add significant build time with many models.

#### Create separate API models and domain models

**Conditional** — use in large apps.

Adds verbosity, but prevents complexity in ViewModels and use-cases.

### App structure

Well organized code benefits both the app and the team.

#### Use dependency injection

**Strongly recommend** — prevents globally accessible objects, making code less error prone. **Modern 2026 approach**: Use Riverpod ( `provider` package is legacy). Riverpod provides compile-time
safety, test overrides, and automatic cleanup—avoid manual `GetIt` singletons.

##### Riverpod Provider Types by Use Case

| Pattern                 | Use Case              | Example                                |
| ----------------------- | --------------------- | -------------------------------------- |
| `FutureProvider`        | Load async data once  | Fetch user profile on screen open      |
| `StateNotifierProvider` | Complex state + logic | Authentication state with login/logout |
| `Provider.family`       | Parameterized DI      | Repositories scoped to projectId       |
| `.autoDispose`          | Auto-cleanup          | Feature-scoped providers               |

**Service-Repository-UseCase Chain** (via Riverpod):

```dart

// ✅ CORRECT: Riverpord does dependency injection
final diagramFileServiceProvider = Provider.family<DiagramFileService, String>(
  (ref, projectRoot) => DiagramFileService(projectRoot),
);

final diagramRepositoryProvider = Provider.family<DiagramRepository, String>(
  (ref, projectRoot) {
    final fileService = ref.watch(diagramFileServiceProvider(projectRoot));
    return DiagramRepository(fileService: fileService);
  },
);

// ❌ WRONG: Manual GetIt (legacy)
GetIt.I.registerSingleton<DiagramRepository>(repo);
```

**Anti-Pattern**: Repositories must not call other repositories; orchestrate dependencies in UseCases instead:

```dart

// ❌ WRONG: Cross-repository call
class UserRepository {
  Future<User> getUserWithProfile(id) {
    final user = await authRepository.getUser(id);  // VIOLATION
    return user;
  }
}

// ✅ CORRECT: Use UseCase for orchestration
class GetUserWithProfileUseCase {
  GetUserWithProfileUseCase(this.userRepo, this.profileRepo);
  Future<User> call(id) {
    final user = await userRepo.getUser(id);
    final profile = await profileRepo.getProfile(id);
    return user.copyWith(profile: profile);
  }
}
```

#### Use `go_router` for navigation

**Recommend** — preferred for ~90% of Flutter apps. For use-cases it doesn't solve, use the `Flutter Navigator API` directly or other `pub.dev` packages.

#### Use standardized naming conventions for classes, files and directories

**Recommend** — name classes for the architectural component they represent, e.g. `HomeViewModel` , `HomeScreen` , `UserRepository` , `ClientApiService` . Avoid names confusable with Flutter SDK
objects - e.g. put shared widgets in `ui/core/` , not `/widgets` .

#### Use feature-based folder organization (NOT layer-based)

**Strongly recommend** — organize by feature, not by architectural layer. Each feature owns its own UI, business logic, and data handling:

```text

lib/
├── features/
│   ├── auth/
│   │   ├── data/
│   │   │   ├── datasources/
│   │   │   ├── models/
│   │   │   └── repositories/
│   │   ├── domain/
│   │   │   ├── entities/
│   │   │   ├── repositories/ (interfaces)
│   │   │   └── usecases/
│   │   └── presentation/
│   │       ├── pages/
│   │       ├── widgets/
│   │       └── providers/
├── core/
│   ├── config/
│   ├── constants/
│   └── utils/
└── shared/
    ├── widgets/
    └── themes/

```

**Benefits**: Single feature folder when modifying, parallel team development, easy feature deletion, clear boundaries.

#### Use module isolation for large teams

**Conditional** — use only if: 5+ developers, independent feature delivery, or long-term maintenance required. Skip for solo dev, MVPs, or projects <50 screens.

Each module is self-contained: own UI, business logic, data handling. No shared state across modules. Centralize routing at the app layer.

#### Use abstract repository classes

**Strongly recommend** — repositories are the source of truth for app data and handle external API communication.
Abstract classes allow different implementations per environment (development, staging).

### Testing

Good testing practices keep the app flexible and make adding new logic/UI low risk.

#### Test architectural components separately, and together

**Strongly recommend** — cover the unit and widget levels:

- Unit tests for every service, repository, and ViewModel class - test each method's logic individually
- Widget tests for views - routing and dependency injection matter most

#### Make fakes for testing (and write code that takes advantage of fakes.)

**Strongly recommend** — fakes focus on inputs/outputs, not a method's inner workings - writing with this in mind forces modular, lightweight functions/classes with well-defined inputs and outputs.
