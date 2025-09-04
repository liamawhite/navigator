// Copyright 2025 Navigator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import { useState } from 'react';

/**
 * Custom hook for persisting state in localStorage
 * @param key - The localStorage key to use
 * @param defaultValue - The default value to use if no stored value exists
 * @param validator - Optional function to validate stored data structure
 * @returns A tuple of [value, setValue] similar to useState
 */
export function useLocalStorage<T>(
    key: string,
    defaultValue: T,
    validator?: (data: unknown) => data is T
): [T, (value: T | ((prev: T) => T)) => void] {
    // Prefix all keys to avoid conflicts with other applications
    const prefixedKey = `navigator-${key}`;

    // Initialize state from localStorage or default value
    const [storedValue, setStoredValue] = useState<T>(() => {
        try {
            const item = window.localStorage.getItem(prefixedKey);
            if (!item) return defaultValue;

            const parsedItem = JSON.parse(item);

            // Validate stored data structure if validator is provided
            if (validator && !validator(parsedItem)) {
                console.warn(
                    'Invalid data structure in localStorage key: %s:',
                    prefixedKey,
                    parsedItem
                );
                // Remove invalid data
                window.localStorage.removeItem(prefixedKey);
                return defaultValue;
            }

            return parsedItem;
        } catch (error) {
            console.warn(
                'Failed to parse localStorage key:', prefixedKey, error
            );
            // Remove corrupted data
            window.localStorage.removeItem(prefixedKey);
            return defaultValue;
        }
    });

    // Return a wrapped version of useState's setter function that persists to localStorage
    const setValue = (value: T | ((prev: T) => T)) => {
        try {
            // Allow value to be a function so we have the same API as useState
            const valueToStore =
                value instanceof Function ? value(storedValue) : value;

            // Save to local storage
            window.localStorage.setItem(
                prefixedKey,
                JSON.stringify(valueToStore)
            );

            // Save state
            setStoredValue(valueToStore);
        } catch (error) {
            console.warn('Error setting localStorage key:', prefixedKey, error);
            // Still update the state even if localStorage fails
            const valueToStore =
                value instanceof Function ? value(storedValue) : value;
            setStoredValue(valueToStore);
        }
    };

    return [storedValue, setValue];
}
