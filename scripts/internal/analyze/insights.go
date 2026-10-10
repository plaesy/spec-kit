package analyze

// aiInsights mirrors the JSON shape produced by generate_ai_insights().
type aiInsights struct {
	Overview        string   `json:"overview"`
	Recommendations []string `json:"recommendations"`
}

// generateAIInsights mirrors generate_ai_insights(): a big switch over the
// primary detected project type, each case producing a short overview plus
// a handful of framework-specific recommendations, followed by one
// universal recommendation when the project is very small.
func generateAIInsights(projectType string, totalFiles int) aiInsights {
	var overview string
	var recs []string

	switch projectType {
	case "nextjs":
		overview = "Next.js application with React framework featuring SSR/SSG capabilities"
		recs = []string{
			"Implement comprehensive SEO optimization with metadata API",
			"Add performance monitoring with Next.js Analytics",
			"Consider incremental static regeneration for dynamic content",
		}
	case "react":
		overview = "React application with component-based architecture and hooks system"
		recs = []string{
			"Consider adding TypeScript for better type safety",
			"Implement comprehensive unit tests with Jest and React Testing Library",
			"Add state management solution like Redux Toolkit or Zustand",
		}
	case "vue":
		overview = "Vue.js application with reactive component system and progressive framework"
		recs = []string{
			"Consider Vue 3 Composition API for better code organization",
			"Add TypeScript support with vue-tsc",
			"Implement Vue Router for single page applications",
		}
	case "nuxt":
		overview = "Nuxt.js application providing Vue.js framework with SSR capabilities"
		recs = []string{
			"Optimize for search engines with proper meta tags",
			"Implement static site generation for better performance",
			"Add Nuxt modules for enhanced functionality",
		}
	case "svelte":
		overview = "Svelte application with compile-time optimizations and reactive components"
		recs = []string{
			"Add TypeScript support for better development experience",
			"Implement SvelteKit for full-stack applications",
			"Consider adding unit tests with Vitest",
		}
	case "astro":
		overview = "Astro application focused on content-rich sites with minimal JavaScript"
		recs = []string{
			"Optimize for performance with island architecture",
			"Implement proper SEO and meta management",
			"Add Markdown/MDX content processing capabilities",
		}
	case "remix":
		overview = "Remix web framework with focus on web fundamentals and nested routes"
		recs = []string{
			"Implement proper error boundaries and loading states",
			"Add database integration with Prisma or similar",
			"Optimize for progressive enhancement",
		}
	case "gatsby":
		overview = "Gatsby static site generator with React and GraphQL data layer"
		recs = []string{
			"Optimize build performance with Gatsby caching",
			"Implement proper image optimization with gatsby-plugin-image",
			"Add SEO optimization with gatsby-plugin-react-helmet",
		}
	case "express":
		overview = "Express.js web application with minimal Node.js framework"
		recs = []string{
			"Add comprehensive error handling middleware",
			"Implement proper logging with Winston or Morgan",
			"Consider adding TypeScript support",
		}
	case "nestjs":
		overview = "NestJS application with TypeScript and architectural patterns"
		recs = []string{
			"Implement comprehensive validation with class-validator",
			"Add OpenAPI/Swagger documentation",
			"Consider microservices architecture with NestJS modules",
		}
	case "vite":
		overview = "Vite-powered development environment with fast build tooling"
		recs = []string{
			"Configure proper build optimization for production",
			"Add plugin ecosystem for enhanced functionality",
			"Implement HMR optimization for development workflow",
		}
	case "django":
		overview = "Django web application with batteries-included framework"
		recs = []string{
			"Add comprehensive API documentation with Django REST framework",
			"Implement proper caching strategies with Redis",
			"Add background task processing with Celery",
		}
	case "fastapi":
		overview = "FastAPI application with high-performance async web framework"
		recs = []string{
			"Add comprehensive OpenAPI/Swagger documentation",
			"Implement dependency injection for better testing",
			"Add async database operations with SQLAlchemy 2.0",
		}
	case "flask":
		overview = "Flask web application with lightweight and extensible framework"
		recs = []string{
			"Add SQLAlchemy for database management",
			"Implement proper application factory pattern",
			"Add Flask extensions for enhanced functionality",
		}
	case "poetry":
		overview = "Python project managed with Poetry for dependency management"
		recs = []string{
			"Configure proper dependency groups for development",
			"Add comprehensive pytest configuration",
			"Implement pre-commit hooks with poetry hooks",
		}
	case "pipenv":
		overview = "Python project with Pipenv for virtual environment and dependency management"
		recs = []string{
			"Configure proper Pipfile for development dependencies",
			"Add pipenv scripts for common tasks",
		}
	case "flutter":
		overview = "Flutter application with cross-platform mobile development framework"
		recs = []string{
			"Add comprehensive unit tests with flutter test",
			"Add effective linter configuration with flutter analyze",
			"Consider adding integration tests for critical user flows",
		}
	case "gin":
		overview = "Gin web framework for Go with high-performance HTTP router"
		recs = []string{
			"Add comprehensive middleware for authentication and logging",
			"Implement proper error handling and validation",
			"Add database integration with GORM or sqlx",
		}
	case "echo":
		overview = "Echo web framework for Go with high performance and extensibility"
		recs = []string{
			"Implement proper middleware stack",
			"Add comprehensive error handling",
			"Consider adding database integration",
		}
	case "fiber":
		overview = "Fiber web framework for Go inspired by Express.js"
		recs = []string{
			"Add comprehensive middleware configuration",
			"Implement proper error handling",
			"Add database integration for full-stack applications",
		}
	case "actix":
		overview = "Actix Web framework for Rust with high-performance actor model"
		recs = []string{
			"Implement proper middleware for authentication",
			"Add comprehensive error handling",
			"Consider adding database integration with sqlx",
		}
	case "rocket":
		overview = "Rocket web framework for Rust with type-safe and fast development"
		recs = []string{
			"Implement proper database integration with Diesel",
			"Add comprehensive testing strategies",
			"Configure proper async/await patterns",
		}
	case "axum":
		overview = "Axum web framework for Rust built on Tokio with async support"
		recs = []string{
			"Implement proper async database integration",
			"Add comprehensive middleware for logging and tracing",
			"Consider adding tower middleware ecosystem",
		}
	case "rust":
		overview = "Rust application with focus on memory safety and performance"
		recs = []string{
			"Add comprehensive unit and integration tests",
			"Implement proper error handling with Result types",
			"Consider adding async runtime optimization",
		}
	case "springboot":
		overview = "Spring Boot application with enterprise-grade framework"
		recs = []string{
			"Add comprehensive Spring Security configuration",
			"Implement proper database integration with Spring Data JPA",
			"Add RESTful API documentation with OpenAPI",
		}
	case "spring":
		overview = "Spring framework application with dependency injection and AOP"
		recs = []string{
			"Configure proper Spring context management",
			"Add comprehensive testing with Spring Test",
			"Implement proper MVC patterns",
		}
	case "maven":
		overview = "Java project managed with Maven build system"
		recs = []string{
			"Configure proper dependency management",
			"Add comprehensive plugin configuration",
			"Implement proper build lifecycle",
		}
	case "gradle":
		overview = "Java project managed with Gradle build system"
		recs = []string{
			"Configure proper build scripts with Kotlin DSL",
			"Add comprehensive plugin ecosystem",
			"Implement proper dependency management",
		}
	case "ktor":
		overview = "Ktor framework for Kotlin with asynchronous web applications"
		recs = []string{
			"Add comprehensive structured logging",
			"Implement proper content negotiation",
			"Add database integration with Exposed",
		}
	case "rails":
		overview = "Ruby on Rails application with convention over configuration"
		recs = []string{
			"Add comprehensive testing with RSpec or Minitest",
			"Implement proper background jobs with Sidekiq",
			"Add API versioning strategies",
		}
	case "sinatra":
		overview = "Sinatra DSL for Ruby with lightweight web applications"
		recs = []string{
			"Add comprehensive middleware configuration",
			"Implement proper RESTful API patterns",
			"Consider adding ActiveRecord for database",
		}
	case "laravel":
		overview = "Laravel PHP framework with elegant syntax and comprehensive ecosystem"
		recs = []string{
			"Add comprehensive API documentation with Laravel API resources",
			"Implement proper queue system with Redis and Horizon",
			"Add comprehensive testing with PHPUnit and Laravel Dusk",
		}
	case "symfony":
		overview = "Symfony PHP framework with reusable components and MVC architecture"
		recs = []string{
			"Configure proper service container and dependency injection",
			"Add comprehensive form validation and security",
			"Implement proper API platform integration",
		}
	case "wordpress":
		overview = "WordPress content management system with plugin and theme development"
		recs = []string{
			"Implement proper WordPress coding standards",
			"Add comprehensive security practices",
			"Consider adding custom post types and taxonomies",
		}
	case "aspnet":
		overview = "ASP.NET Core application with cross-platform web framework"
		recs = []string{
			"Add comprehensive authentication and authorization",
			"Implement proper dependency injection patterns",
			"Add Entity Framework Core for database operations",
		}
	case "dotnet":
		overview = ".NET application with managed runtime and comprehensive framework"
		recs = []string{
			"Configure proper NuGet package management",
			"Add comprehensive unit testing with xUnit or NUnit",
			"Implement proper async/await patterns",
		}
	case "swift":
		overview = "Swift application with safe and performant programming language"
		recs = []string{
			"Add comprehensive unit tests with XCTest",
			"Implement proper error handling with Result types",
			"Consider adding async/await patterns",
		}
	case "docker":
		overview = "Docker containerized application with containerization platform"
		recs = []string{
			"Add multi-stage builds for optimization",
			"Implement proper health checks",
			"Consider adding Docker Compose for orchestration",
		}
	case "docker-compose":
		overview = "Docker Compose application with multi-container orchestration"
		recs = []string{
			"Configure proper networking and volumes",
			"Add environment-specific configurations",
			"Implement proper service dependencies",
		}
	case "terraform":
		overview = "Terraform infrastructure as code with cloud resource management"
		recs = []string{
			"Implement proper state management strategies",
			"Add comprehensive variable validation",
			"Configure proper remote state backends",
		}
	case "vagrant":
		overview = "Vagrant development environment with virtualization management"
		recs = []string{
			"Configure proper provisioning scripts",
			"Add multi-machine setup if needed",
			"Implement proper network configuration",
		}
	case "kustomize":
		overview = "Kustomize Kubernetes configuration management with template-free customization"
		recs = []string{
			"Configure proper overlay structure",
			"Add comprehensive resource organization",
			"Implement proper secret management",
		}
	case "spec-kit":
		overview = "Plaesy Spec-Kit framework for AI-assisted development"
		recs = []string{"Add more AI platform integrations"}
	default:
		overview = "Generic project structure"
		recs = []string{
			"Add README.md with project documentation",
			"Consider adding automated testing",
			"Implement proper version control practices",
		}
	}

	if totalFiles < 10 {
		recs = append(recs, "Expand project with additional features and modules")
	}

	return aiInsights{Overview: overview, Recommendations: recs}
}
