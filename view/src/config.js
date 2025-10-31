/**
 * Application configuration
 */

// API base URL - defaults to localhost:8080 for development
// Can be overridden via environment variable VITE_API_BASE_URL
export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'
