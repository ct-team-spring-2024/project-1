# Project README

## Team Members
- **Ali Rasekhi**
- **Mohammad Broughani**
- **Amir Ahmad Shafiee**

---

## Project Overview

This project is structured to separate concerns and ensure modularity, making it easier to maintain and extend. Below is a breakdown of the key components and their functionalities:

### 1. **Main Logic**
The core logic of the application resides in the **`internal` package**. This package contains the essential algorithms and business logic that drive the functionality of the project. If you're looking to understand or modify the fundamental operations of the application, this is where you should start.

### 2. **Network Functionality**
All network-related operations are encapsulated in the **`pkg/network/downloader`** package. This includes tasks such as downloading files, handling network requests, and managing connections. The separation of network logic ensures that the codebase remains clean and focused, with all networking concerns isolated in this module.

### 3. **Test Scenarios**
The **`cmd/idm/main.go`** file contains several test scenarios designed to validate the functionality of the application. These scenarios are useful for testing various edge cases and ensuring that the application behaves as expected under different conditions. Developers can use these tests as a reference for understanding how the system is intended to work.

### 4. **UI Features**
The **`ui/main_ui.go`** file implements a few user interface features. While the UI currently includes only a limited set of functionalities, the underlying logic implemented in other parts of the project is far more extensive. This separation allows for future expansion of the UI without affecting the core logic.

---

## How to Get Started

1. Clone the repository.
2. Navigate to the root directory of the project.
3. Explore the `internal` package to understand the main logic.
4. Check out the `pkg/network/downloader` for network-related functionality.
5. Run the test scenarios in `cmd/idm/main.go` to verify the application's behavior.
6. Review the UI features in `ui/main_ui.go` and consider extending them based on the robust logic already implemented.

---

## Contribution Guidelines

We welcome contributions from team members and collaborators. When contributing, please adhere to the following guidelines:
- Ensure your changes align with the modular structure of the project.
- Add relevant tests in `cmd/idm/main.go` to cover new functionality.
- Document any new features or significant changes in the code.

---
