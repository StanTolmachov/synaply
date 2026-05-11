## Instructions

- Отвечай на русском языке независимо от языка промта если нет иной команды.

## Scope

- Эти правила действуют для всего репозитория.
- Если в подпапке есть свой `AGENTS.md`, он имеет приоритет для этой подпапки.

## Code Style

- Следовать существующему стилю кода в файле/пакете.
- Не добавлять новые зависимости без явной необходимости.
- Не делать массовые рефакторы вне задачи.

## Change Rules

- Изменения должны быть минимальными и целевыми.
- Не трогать несвязанные файлы.
- Сохранять обратную совместимость публичных интерфейсов, если это не оговорено в задаче.

## Safety

- Никогда не коммитить секреты, токены, приватные ключи.
- Не менять CI/CD, infra, миграции и права доступа без явного запроса.

## Go Best Practices

- Prefer stdlib over external dependencies.
- Keep interfaces small and consumer-defined.
- Use table-driven tests.
- Avoid premature abstractions.
- Use context.Context as first parameter where applicable.
- Wrap errors with %w.
- Prefer composition over inheritance-like patterns.
- Keep packages focused and small.
- Avoid global state.
- Use slog for logging.
- Follow existing project architecture.

## AI Coding Rules

- Before changing code, inspect neighboring files and patterns.
- Keep diffs minimal.
- Do not introduce new abstractions unless necessary.
- Do not rewrite working code.
- Preserve backward compatibility.
- Prefer readability over cleverness.

