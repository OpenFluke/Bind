# Bind

![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)
![Build Status](https://img.shields.io/github/workflow/status/openneuralforge/bind/Build)

## Overview

**Bind** is a pivotal component of the [Neural Forge](https://github.com/openfluke) ecosystem, engineered to seamlessly integrate the `Blueprint` Go package with JavaScript environments through WebAssembly (WASM). By leveraging reflection-based method wrapping, Bind dynamically exposes Go methods to the web, enabling efficient execution of AI framework functionalities directly within browser environments. This facilitates interactive testing, benchmarking, and paves the way for future enhancements utilizing technologies like WebGPU for distributed evaluation.

## Features

- **Dynamic Method Introspection**: Automatically detects and wraps all available methods in the `Blueprint` package for browser-based execution.
- **Browser-Based Execution**: Executes `Blueprint` AI computations directly in web browsers via WebAssembly.
- **Self-Updating Environment**: Automatically includes new or modified methods from the `Blueprint` framework in each build.
- **Test, Benchmark, and Showcase**: Easily demonstrate `Blueprint`'s capabilities with real-time, browser-accessible results.

## Getting Started

### Option 1: Using npm (Node.js)

1. **Install the Package**:
   ```bash
   npm install @openneuralforge/bind@latest
   ```

2. **Create a Node.js Script**:
   Create a new file (e.g., `index.js`) with the following content:
   ```javascript
   import fs from 'fs';
   import path from 'path';
   import { fileURLToPath } from 'url';
   import vm from 'vm';

   // Resolve file paths
   const __filename = fileURLToPath(import.meta.url);
   const __dirname = path.dirname(__filename);
   const wasmExecPath = path.join(__dirname, 'node_modules/@openneuralforge/bind/dist/wasm_exec.js');
   const wasmPath = path.join(__dirname, 'node_modules/@openneuralforge/bind/dist/main.wasm');

   // Load the wasm_exec.js file into the global context
   const wasmExecCode = fs.readFileSync(wasmExecPath, 'utf8');
   vm.runInThisContext(wasmExecCode);
   const go = new Go(); // The Go constructor is now available globally

   // Function to load and initialize the WebAssembly module
   async function loadWasm() {
     const wasmBuffer = fs.readFileSync(wasmPath);
     const wasmModule = await WebAssembly.instantiate(wasmBuffer, go.importObject);
     go.run(wasmModule.instance);
     return wasmModule.instance;
   }

   // Main async function to initialize and test WASM
   (async () => {
     try {
       console.log("Loading WASM...");
       const wasmInstance = await loadWasm();
       console.log("WASM loaded successfully!");

       // Example of calling a WASM function
       if (typeof globalThis.GetBlueprintMethodsJSON === "function") {
         const methods = globalThis.GetBlueprintMethodsJSON();
         console.log("Available methods:", methods);
       } else {
         console.log("GetBlueprintMethodsJSON function is not defined in WASM.");
       }
     } catch (error) {
       console.error("Error loading WASM:", error);
     }
   })();
   ```

### Option 2: Building from Source

1. **Clone the Repositories**:
   Clone the LayerForge AI framework into the parent folder:
   ```bash
   git clone https://github.com/openfluke/Anvil.git
   ```

   Then, clone the bind repository inside the same parent directory:
   ```bash
   git clone https://github.com/openfluke/bind.git
   ```

2. **Build the WASM File**:
   Navigate to the `bind` directory, then run the following command to compile `LayerForge` into WebAssembly:
   ```bash
   GOOS=js GOARCH=wasm go build -o blueprint.wasm main.go
   ```

3. **Prepare the WASM Execution Environment**:
   Ensure `wasm_exec.js` (found in your Go installation, typically under `GOROOT/misc/wasm/wasm_exec.js`) is in the same directory as your `index.html` file.

4. **Start a Local HTTP Server**:
   To serve the files locally, start an HTTP server from within the `bind` directory:
   ```bash
   python3 -m http.server 8000
   ```

5. **Access the Interface**:
   Open your browser and navigate to:
   ```
   http://localhost:8000
   ```
   You should see the interface where you can:
   - Run introspection on the LayerForge framework to view available methods
   - Execute benchmarks and other LayerForge methods directly in your browser

## Technical Architecture

Bind employs advanced techniques to ensure robust and flexible interoperability between Go and JavaScript. The core functionalities are centered around reflection-based method wrapping, JavaScript interoperability, and WebAssembly integration.

### Reflection-Based Method Wrapping

At the heart of Bind lies the **reflection-based method wrapping** mechanism. This approach utilizes Go's `reflect` package to dynamically inspect and wrap methods from the `Blueprint` package at runtime.

- **Dynamic Method Access**: Bind introspects the `Blueprint` struct to identify and access all available methods without the need for manual binding definitions.
- **Runtime Wrapping**: Each method is encapsulated within a `js.Func`, enabling it to be invoked directly from JavaScript.
- **Parameter Handling**: Bind intelligently parses and maps JSON-encoded inputs from JavaScript to the appropriate Go types, ensuring type safety and correctness during method invocation.
- **Result Serialization**: The results from Go methods are serialized back into JSON strings, allowing seamless consumption within JavaScript environments.

### JavaScript Interoperability

Bind establishes a robust interoperability layer between Go and JavaScript, enabling bi-directional communication and method invocation.

- **Exported Bindings**: All methods from the `Blueprint` package are exposed as global JavaScript functions, making them readily accessible within the browser context.
- **Serialization Mechanism**: Inputs and outputs are handled through JSON serialization and deserialization, facilitating smooth data exchange between the two languages.
- **Error Handling**: Bind ensures that any errors during method invocation are gracefully captured and communicated back to JavaScript, providing meaningful feedback for debugging and resilience.

### WebAssembly Integration

By compiling Go code into WebAssembly, Bind ensures high-performance execution of AI framework functionalities directly within web browsers.

- **Efficient Execution**: WASM offers near-native performance, allowing complex AI operations to run efficiently on the client side without significant latency.
- **Portability**: The WASM module generated by Bind is platform-agnostic, ensuring consistent behavior across different browsers and devices.
- **Future-Proof Design**: The architecture is designed to accommodate future integrations, such as leveraging WebGPU for enhanced graphical computations and distributed evaluations, further expanding the capabilities of the AI framework in web environments.

## Component Breakdown

1. **Method Wrapper Function (`methodWrapper`)**:
   - Utilizes reflection to access methods of the `Blueprint` struct.
   - Wraps each method into a `js.Func` for exposure to JavaScript.
   - Handles input parsing and output serialization.

2. **Main Execution Flow (`main`)**:
   - Instantiates the `Blueprint` struct.
   - Retrieves all methods via reflection-based introspection.
   - Exposes each method as a global JavaScript function.
   - Keeps the WASM module active to listen for JavaScript invocations.

## Design Principles

- **Modularity**: Bind is designed as an independent repository, ensuring that it can be maintained and developed separately from other components like `Anvil` and `Hammer`.
- **Scalability**: The reflection-based approach allows Bind to automatically adapt to changes in the `Blueprint` package, reducing the maintenance overhead as the AI framework evolves.
- **Performance**: By leveraging WASM and efficient serialization mechanisms, Bind ensures that method invocations are executed with minimal overhead, maintaining high performance within web environments.
- **Extensibility**: The architecture supports future enhancements, such as integrating WebGPU, enabling more complex and distributed AI computations directly in the browser.