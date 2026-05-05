import { emitter } from './event-emitter.js'

const API_BASE = 'http://localhost:3000/api/v1';

// Headers management
const defaultHeaders = {
    'Content-Type': 'application/json',
};

// Common Pattern: do not export each function; instead put them in
// an object and export only the object
export const DataService = {
    // Set custom headers (e.g., auth token)
    setHeaders(customHeaders) {
        Object.assign(defaultHeaders, customHeaders);
    },

    // Set auth token specifically
    setAuthToken(token) {
        if (token) {
            defaultHeaders['Authorization'] = `Bearer ${token}`;
        } else {
            delete defaultHeaders['Authorization'];
        }
    },

    // Get current headers
    getHeaders() {
        return { ...defaultHeaders };
    },

    // Build fetch options with headers
    buildFetchOptions(method = 'GET', body = null) {
        const options = {
            method,
            headers: this.getHeaders(),
        };
        if (body) {
            options.body = JSON.stringify(body);
        }
        return options;
    },

    async fetchServices(page = 1, pageSize = 5) {
        // 1. announce that loading has started
        emitter.emit('services:loading');
        try {
            // Build URL with query parameters
            const params = new URLSearchParams();
            params.append('page', page);
            params.append('page_size', pageSize);
            const url = `${API_BASE}/services?${params.toString()}`;
            
            const res  = await fetch(url, this.buildFetchOptions('GET'));
            if(!res.ok) {
                throw new Error(`Server error: ${res.status}`);
            }
            const data = await res.json();
            // Extract services array and metadata from response
            const services = data.services || [];
            const metadata = data.metadata || {};
            // 2. announce success — pass both services and metadata
            emitter.emit('services:loaded', { services, metadata });
        } catch(err) {
            // 3. announce failure — pass the error message
            console.error('Error fetching services:', err);
            emitter.emit('services:error', err.message);
        }
    },

    async createService(payload) {
        emitter.emit('services:loading');
        try {
            const res = await fetch(`${API_BASE}/services`, this.buildFetchOptions('POST', payload));
            if (!res.ok) { 
                throw new Error(`Server error: ${res.status}`);
            }

            const newService = await res.json();
            emitter.emit('services:created', newService);
        } catch (err) {
            emitter.emit('services:error', err.message);
        }
    }
}
