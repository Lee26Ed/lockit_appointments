function createEmitter() {
    const listeners = new Map();

    function on(event, callback)  {
        if(!listeners.has(event)) {
            listeners.set(event, []);
        }
        listeners.get(event).push(callback)
    }
    function emit(event, data) {
        if (!listeners.has(event)) {
            return;
        }
        listeners.get(event).forEach(cb => cb(data));
    }
    function off(event, callback) {
        if (!listeners.has(event)) {
            return;
        }
        listeners.set(event, listeners.get(event)
                    .filter(cb => cb !== callback));
    }

    return { on, emit, off };
}

// One shared instance for the whole app
// emitter has access to all three functions - on, emit, off
export const emitter = createEmitter();