// Package mocks provides hand-rolled mock implementations of all service interfaces
// used in the paso application. These mocks are designed for use in both CLI and TUI
// unit tests, enabling fast, isolated testing without database dependencies.
//
// Each mock follows the same pattern established by internal/services/taskevent/mock.go:
//
//   - Thread-safe: All mocks use sync.Mutex to protect concurrent access to recorded calls
//     and injected return values, making them safe for use in parallel test scenarios.
//
//   - Call recording: Every method call is recorded with its method name, arguments, and
//     a primary ID (where applicable). Tests can verify which methods were called, how many
//     times, and with what arguments.
//
//   - Per-method error/result injection: Each method has corresponding exported fields
//     (e.g., GetTaskDetailErr, GetTaskDetailResult) that allow tests to configure return
//     values before exercising the code under test.
//
//   - Standard helpers: All mocks provide Reset(), GetCalls(), HasCall(), and CallCount()
//     methods for test assertions.
//
//   - Compile-time interface verification: Each mock includes a compile-time check like
//     var _ Interface = (*MockType)(nil) to ensure the mock always satisfies its interface.
//
// These files are intentionally regular .go files (not _test.go) to allow cross-package
// imports. This is necessary because Go's _test.go files cannot be imported by other
// packages. The minor binary size increase (~50-100KB) is acceptable for a CLI tool.
//
// Available mocks:
//
//   - MockEventPublisher: implements events.EventPublisher (6 methods)
//   - MockTaskService: implements task.Service (31 methods across 6 sub-interfaces)
//   - MockTaskEventService: implements taskevent.Service (9 methods) - for cross-package tests
//   - MockProjectService: implements project.Service (7 methods)
//   - MockColumnService: implements column.Service (8 methods)
//   - MockLabelService: implements label.Service (6 methods)
//   - MockAssigneeService: implements assignee.Service (6 methods)
//   - MockGitDetector: implements git.Detector (3 methods)
package mocks
