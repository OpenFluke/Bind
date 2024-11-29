export async function loadWasm(wasmPath = './main.wasm') {
    const go = new Go();
    const wasmModule = await WebAssembly.instantiateStreaming(fetch(wasmPath), go.importObject);
    go.run(wasmModule.instance);
    return globalThis;
}
