---
description: "Instructions for writing Dart and Flutter code following the official recommendations."
applyTo: "**/*.dart, **/*.yaml, **/*.yml, pubspec.yaml"
---

# Dart and Flutter

Best practices from the Dart/Flutter teams, taken from [Effective Dart](https://dart.dev/effective-dart) and [Architecture Recommendations](https://docs.flutter.dev/app-architecture/recommendations).

## Effective Dart

Two overarching themes:

1. **Be consistent.** Formatting/casing debates are subjective, but consistency itself is objectively helpful - code that looks different should be different for a meaningful reason.
2. **Be brief.** Prefer the most concise way to express intent - economical, not dense (not code-golf).

### The topics

- **Style** - layout/organization rules `dart format` doesn't handle, plus identifier casing (`camelCase`, `using_underscores`)
- **Documentation** - what goes inside doc comments and regular comments
- **Usage** - how to use language features to implement behavior
- **Design** - designing consistent, usable library APIs

### How to read the topics

Each guideline starts with:

- **DO** - always follow, essentially no valid exception
- **DON'T** - almost never a good idea
- **PREFER** - should follow, but exceptions can make sense - understand the implications before ignoring
- **AVOID** - shouldn't do, but rare good reasons may exist
- **CONSIDER** - may or may not apply depending on circumstances/preference

Some guidelines list explicit exceptions (not necessarily exhaustive - use judgement). Goal throughout: readable, maintainable code.

### Rules

#### Style

##### Identifiers

- DO name types using `UpperCamelCase`.
- DO name extensions using `UpperCamelCase`.
- DO name packages, directories, and source files using `lowercase_with_underscores`.
- DO name import prefixes using `lowercase_with_underscores`.
- DO name other identifiers using `lowerCamelCase`.
- PREFER using `lowerCamelCase` for constant names.
- DO capitalize acronyms and abbreviations longer than two letters like words.
- PREFER using wildcards for unused callback parameters.
- DON'T use a leading underscore for identifiers that aren't private.
- DON'T use prefix letters.
- DON'T explicitly name libraries.

##### Ordering

- DO place `dart:` imports before other imports.
- DO place `package:` imports before relative imports.
- DO specify exports in a separate section after all imports.
- DO sort sections alphabetically.

##### Formatting

- DO format your code using `dart format`.
- CONSIDER changing your code to make it more formatter-friendly.
- PREFER lines 80 characters or fewer.
- DO use curly braces for all flow control statements.

#### Documentation

##### Comments

- DO format comments like sentences.
- DON'T use block comments for documentation.

##### Doc comments

- DO use `///` doc comments to document members and types.
- PREFER writing doc comments for public APIs.
- CONSIDER writing a library-level doc comment.
- CONSIDER writing doc comments for private APIs.
- DO start doc comments with a single-sentence summary.
- DO separate the first sentence of a doc comment into its own paragraph.
- AVOID redundancy with the surrounding context.
- PREFER starting comments of a function or method with third-person verbs if its main purpose is a side effect.
- PREFER starting a non-boolean variable or property comment with a noun phrase.
- PREFER starting a boolean variable or property comment with "Whether" followed by a noun or gerund phrase.
- PREFER a noun phrase or non-imperative verb phrase for a function or method if returning a value is its primary purpose.
- DON'T write documentation for both the getter and setter of a property.
- PREFER starting library or type comments with noun phrases.
- CONSIDER including code samples in doc comments.
- DO use square brackets in doc comments to refer to in-scope identifiers.
- DO use prose to explain parameters, return values, and exceptions.
- DO put doc comments before metadata annotations.

##### Markdown

- AVOID using markdown excessively.
- AVOID using HTML for formatting.
- PREFER backtick fences for code blocks.

##### Writing

- PREFER brevity.
- AVOID abbreviations and acronyms unless they are obvious.
- PREFER using "this" instead of "the" to refer to a member's instance.

#### Usage

##### Libraries

- DO use strings in `part of` directives.
- DON'T import libraries that are inside the `src` directory of another package.
- DON'T allow an import path to reach into or out of `lib`.
- PREFER relative import paths.

##### Null

- DON'T explicitly initialize variables to `null`.
- DON'T use an explicit default value of `null`.
- DON'T use `true` or `false` in equality operations.
- AVOID `late` variables if you need to check whether they are initialized.
- CONSIDER type promotion or null-check patterns for using nullable types.

##### Strings

- DO use adjacent strings to concatenate string literals.
- PREFER using interpolation to compose strings and values.
- AVOID using curly braces in interpolation when not needed.

##### Collections

- DO use collection literals when possible.
- DON'T use `.length` to see if a collection is empty.
- AVOID using `Iterable.forEach()` with a function literal.
- DON'T use `List.from()` unless you intend to change the type of the result.
- DO use `whereType()` to filter a collection by type.
- DON'T use `cast()` when a nearby operation will do.
- AVOID using `cast()`.

##### Functions

- DO use a function declaration to bind a function to a name.
- DON'T create a lambda when a tear-off will do.

##### Variables

- DO follow a consistent rule for `var` and `final` on local variables.
- AVOID storing what you can calculate.

##### Members

- DON'T wrap a field in a getter and setter unnecessarily.
- PREFER using a `final` field to make a read-only property.
- CONSIDER using `=>` for simple members.
- DON'T use `this.` except to redirect to a named constructor or to avoid shadowing.
- DO initialize fields at their declaration when possible.

##### Constructors

- DO use initializing formals when possible.
- DON'T use `late` when a constructor initializer list will do.
- DO use `;` instead of `{}` for empty constructor bodies.
- DON'T use `new`.
- DON'T use `const` redundantly.

##### Error handling

- AVOID catches without `on` clauses.
- DON'T discard errors from catches without `on` clauses.
- DO throw objects that implement `Error` only for programmatic errors.
- DON'T explicitly catch `Error` or types that implement it.
- DO use `rethrow` to rethrow a caught exception.

##### Asynchrony

- PREFER async/await over using raw futures.
- DON'T use `async` when it has no useful effect.
- CONSIDER using higher-order methods to transform a stream.
- AVOID using Completer directly.
- DO test for `Future<T>` when disambiguating a `FutureOr<T>` whose type argument could be `Object`.

#### Design

##### Names

- DO use terms consistently.
- AVOID abbreviations.
- PREFER putting the most descriptive noun last.
- CONSIDER making the code read like a sentence.
- PREFER a noun phrase for a non-boolean property or variable.
- PREFER a non-imperative verb phrase for a boolean property or variable.
- CONSIDER omitting the verb for a named boolean parameter.
- PREFER the "positive" name for a boolean property or variable.
- PREFER an imperative verb phrase for a function or method whose main purpose is a side effect.
- PREFER a noun phrase or non-imperative verb phrase for a function or method if returning a value is its primary purpose.
- CONSIDER an imperative verb phrase for a function or method if you want to draw attention to the work it performs.
- AVOID starting a method name with `get`.
- PREFER naming a method `to...()` if it copies the object's state to a new object.
- PREFER naming a method `as...()` if it returns a different representation backed by the original object.
- AVOID describing the parameters in the function's or method's name.
- DO follow existing mnemonic conventions when naming type parameters.

##### Libraries

- PREFER making declarations private.
- CONSIDER declaring multiple classes in the same library.

##### Classes and mixins

- AVOID defining a one-member abstract class when a simple function will do.
- AVOID defining a class that contains only static members.
- AVOID extending a class that isn't intended to be subclassed.
- DO use class modifiers to control if your class can be extended.
- AVOID implementing a class that isn't intended to be an interface.
- DO use class modifiers to control if your class can be an interface.
- PREFER defining a pure `mixin` or pure `class` to a `mixin class`.

##### Constructors

- CONSIDER making your constructor `const` if the class supports it.

##### Members

- PREFER making fields and top-level variables `final`.
- DO use getters for operations that conceptually access properties.
- DO use setters for operations that conceptually change properties.
- DON'T define a setter without a corresponding getter.
- AVOID using runtime type tests to fake overloading.
- AVOID public `late final` fields without initializers.
- AVOID returning nullable `Future`, `Stream`, and collection types.
- AVOID returning `this` from methods just to enable a fluent interface.

##### Types

- DO type annotate variables without initializers.
- DO type annotate fields and top-level variables if the type isn't obvious.
- DON'T redundantly type annotate initialized local variables.
- DO annotate return types on function declarations.
- DO annotate parameter types on function declarations.
- DON'T annotate inferred parameter types on function expressions.
- DON'T type annotate initializing formals.
- DO write type arguments on generic invocations that aren't inferred.
- DON'T write type arguments on generic invocations that are inferred.
- AVOID writing incomplete generic types.
- DO annotate with `dynamic` instead of letting inference fail.
- PREFER signatures in function type annotations.
- DON'T specify a return type for a setter.
- DON'T use the legacy typedef syntax.
- PREFER inline function types over typedefs.
- PREFER using function type syntax for parameters.
- AVOID using `dynamic` unless you want to disable static checking.
- DO use `Future<void>` as the return type of asynchronous members that do not produce values.
- AVOID using `FutureOr<T>` as a return type.

##### Parameters

- AVOID positional boolean parameters.
- AVOID optional positional parameters if the user may want to omit earlier parameters.
- AVOID mandatory parameters that accept a special "no argument" value.
- DO use inclusive start and exclusive end parameters to accept a range.

##### Equality

- DO override `hashCode` if you override `==`.
- DO make your `==` operator obey the mathematical rules of equality.
- AVOID defining custom equality for mutable classes.
- DON'T make the parameter to `==` nullable.

---

## Flutter Architecture Recommendations

Architecture best practices - treat as recommendations, not rules, and adapt to your app's requirements. Priority reflects how strongly the Flutter team recommends it:

- **Strongly recommend**: always implement for new apps; strongly consider refactoring existing apps unless it fundamentally clashes with the current approach
- **Recommend**: will likely improve your app
- **Conditional**: can improve your app in certain circumstances

### Separation of concerns

Separate your app into UI and data layers; within those, separate logic into classes by responsibility.

#### Use a three-layer architecture (Domain, Data, Presentation).

**Strongly recommend**

The most important architectural principle:

**Domain Layer** (Pure Dart only—NO `import 'package:flutter'`)

- Entities: core domain models
- Use Cases: business logic orchestration
- Repository interfaces: abstract contracts

**Data Layer**

- Repositories: concrete implementations of domain contracts
- Data Sources: API clients, local DB drivers
- DTOs/Models: JSON serialization
- Caching, error handling, retry logic

**Presentation Layer**

- Widgets/Screens: UI components
- State management providers (Riverpod/BLoC)
- UI logic: navigation, theme, lifecycle

**Critical Rule**: Dependencies point INWARD only → Data → Domain ← Presentation (never reverse, never Data → Domain).

#### Use clearly defined data and UI layers.

**Strongly recommend**

Data layer exposes app data and holds most business logic. UI layer displays data and listens for user events, with separate classes for UI logic and widgets.

#### Use the repository pattern in the data layer.

**Strongly recommend**

Isolates data access logic from the rest of the app - an abstraction layer between business logic and storage (databases, APIs, file systems). In practice: Repository classes and Service classes.

#### Use ViewModels and Views in the UI layer. (MVVM)

**Strongly recommend**

Keeps widgets "dumb", making code much less error prone.

#### Use `ChangeNotifiers` and `Listenables` to handle widget updates.

**Conditional**

> State-management choice ultimately comes down to personal preference.

`ChangeNotifier` is part of the Flutter SDK - a convenient way for widgets to observe ViewModel changes.

#### Do not put logic in widgets.

**Strongly recommend**

Logic belongs in ViewModel methods. A view should only contain:

- Simple if-statements to show/hide widgets based on a ViewModel flag/nullable field
- Animation logic the widget must calculate
- Layout logic based on device info (screen size, orientation)
- Simple routing logic

#### Use a domain layer.

**Conditional**

> Use in apps with complex logic requirements.

Only needed when logic complexity crowds ViewModels or logic repeats across them. Useful in very large apps; unnecessary overhead in most.

### Handling data

Handle data with care for code that's easier to understand, less error prone, and free of malformed/unexpected data.

#### Use unidirectional data flow.

**Strongly recommend**

Data updates flow only from data layer to UI layer; UI interactions are sent to the data layer for processing.

#### Use `Commands` to handle events from user interaction.

**Recommend**

Prevents rendering errors and standardizes how the UI layer sends events to the data layer.

#### Use immutable data models.

**Strongly recommend**

Ensures changes occur only in the proper place (usually data/domain layer) - immutable objects require a new instance to reflect change, preventing accidental UI-layer updates and supporting unidirectional data flow.

#### Use freezed or built_value to generate immutable data models.

**Recommend**

Generate common model methods (JSON ser/des, deep equality, copy). Can add significant build time with many models.

#### Create separate API models and domain models.

**Conditional**

> Use in large apps.

Adds verbosity, but prevents complexity in ViewModels and use-cases.

### App structure

Well organized code benefits both the app and the team.

#### Use dependency injection.

**Strongly recommend**

Prevents globally accessible objects, making code less error prone. **Modern 2026 approach**: Use Riverpod (`provider` package is legacy). Riverpod provides compile-time safety, test overrides, and automatic cleanup—avoid manual `GetIt` singletons.

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

#### Use `go_router` for navigation.

**Recommend**

Preferred for ~90% of Flutter apps. For use-cases it doesn't solve, use the `Flutter Navigator API` directly or other `pub.dev` packages.

#### Use standardized naming conventions for classes, files and directories.

**Recommend**

Name classes for the architectural component they represent, e.g. `HomeViewModel`, `HomeScreen`, `UserRepository`, `ClientApiService`. Avoid names confusable with Flutter SDK objects - e.g. put shared widgets in `ui/core/`, not `/widgets`.

#### Use feature-based folder organization (NOT layer-based).

**Strongly recommend**

Organize by feature, not by architectural layer. Each feature owns its own UI, business logic, and data handling:

```
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

#### Use module isolation for large teams.

**Conditional**

> Use only if: 5+ developers, independent feature delivery, or long-term maintenance required. Skip for solo dev, MVPs, or projects <50 screens.

Each module is self-contained: own UI, business logic, data handling. No shared state across modules. Centralize routing at the app layer.

#### Use abstract repository classes

**Strongly recommend**

Repositories are the source of truth for app data and handle external API communication. Abstract classes allow different implementations per environment (development, staging).

### Testing

Good testing practices keep the app flexible and make adding new logic/UI low risk.

#### Test architectural components separately, and together.

**Strongly recommend**

- Unit tests for every service, repository, and ViewModel class - test each method's logic individually
- Widget tests for views - routing and dependency injection matter most

#### Make fakes for testing (and write code that takes advantage of fakes.)

**Strongly recommend**

Fakes focus on inputs/outputs, not a method's inner workings - writing with this in mind forces modular, lightweight functions/classes with well-defined inputs and outputs.
