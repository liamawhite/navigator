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

import { describe, it, expect, beforeEach, jest } from '@jest/globals';
import { renderHook, act } from '@testing-library/react';
import { useLocalStorage } from './useLocalStorage';

// Mock localStorage
const localStorageMock = (() => {
    let store: Record<string, string> = {};

    return {
        getItem: jest.fn((key: string) => {
            return store[key] || null;
        }),
        setItem: jest.fn((key: string, value: string) => {
            store[key] = value.toString();
        }),
        removeItem: jest.fn((key: string) => {
            delete store[key];
        }),
        clear: jest.fn(() => {
            store = {};
        }),
    };
})();

Object.defineProperty(window, 'localStorage', {
    value: localStorageMock,
});

describe('useLocalStorage', () => {
    beforeEach(() => {
        localStorageMock.clear();
        jest.clearAllMocks();
    });

    it('should return default value when localStorage is empty', () => {
        const { result } = renderHook(() =>
            useLocalStorage('test-key', 'default-value')
        );

        expect(result.current[0]).toBe('default-value');
        expect(localStorageMock.getItem).toHaveBeenCalledWith(
            'navigator-test-key'
        );
    });

    it('should return stored value when localStorage has data', () => {
        localStorageMock.setItem(
            'navigator-test-key',
            JSON.stringify('stored-value')
        );

        const { result } = renderHook(() =>
            useLocalStorage('test-key', 'default-value')
        );

        expect(result.current[0]).toBe('stored-value');
    });

    it('should update localStorage when setValue is called', () => {
        const { result } = renderHook(() =>
            useLocalStorage('test-key', 'default-value')
        );

        act(() => {
            result.current[1]('new-value');
        });

        expect(result.current[0]).toBe('new-value');
        expect(localStorageMock.setItem).toHaveBeenCalledWith(
            'navigator-test-key',
            JSON.stringify('new-value')
        );
    });

    it('should support functional updates', () => {
        const { result } = renderHook(() => useLocalStorage('test-key', 10));

        act(() => {
            result.current[1]((prev) => prev + 5);
        });

        expect(result.current[0]).toBe(15);
        expect(localStorageMock.setItem).toHaveBeenCalledWith(
            'navigator-test-key',
            JSON.stringify(15)
        );
    });

    it('should handle complex objects', () => {
        const defaultObject = { count: 0, name: 'test' };
        const { result } = renderHook(() =>
            useLocalStorage('object-key', defaultObject)
        );

        const newObject = { count: 5, name: 'updated' };
        act(() => {
            result.current[1](newObject);
        });

        expect(result.current[0]).toEqual(newObject);
        expect(localStorageMock.setItem).toHaveBeenCalledWith(
            'navigator-object-key',
            JSON.stringify(newObject)
        );
    });

    it('should return default value and remove corrupted localStorage data', () => {
        // Set invalid JSON
        localStorageMock.setItem('navigator-test-key', 'invalid-json{');

        const consoleSpy = jest
            .spyOn(console, 'warn')
            .mockImplementation(() => {});

        const { result } = renderHook(() =>
            useLocalStorage('test-key', 'default-value')
        );

        expect(result.current[0]).toBe('default-value');
        expect(localStorageMock.removeItem).toHaveBeenCalledWith(
            'navigator-test-key'
        );
        expect(consoleSpy).toHaveBeenCalledWith(
            'Failed to parse localStorage key:',
            'navigator-test-key',
            expect.any(Error)
        );

        consoleSpy.mockRestore();
    });

    it('should validate data structure when validator is provided', () => {
        // Store invalid data structure
        localStorageMock.setItem(
            'navigator-test-key',
            JSON.stringify({ invalid: 'structure' })
        );

        const validator = (
            data: unknown
        ): data is { name: string; age: number } => {
            return (
                typeof data === 'object' &&
                data !== null &&
                'name' in data &&
                'age' in data &&
                typeof (data as Record<string, unknown>).name === 'string' &&
                typeof (data as Record<string, unknown>).age === 'number'
            );
        };

        const consoleSpy = jest
            .spyOn(console, 'warn')
            .mockImplementation(() => {});
        const defaultValue = { name: 'default', age: 0 };

        const { result } = renderHook(() =>
            useLocalStorage('test-key', defaultValue, validator)
        );

        expect(result.current[0]).toEqual(defaultValue);
        expect(localStorageMock.removeItem).toHaveBeenCalledWith(
            'navigator-test-key'
        );
        expect(consoleSpy).toHaveBeenCalledWith(
            'Invalid data structure in localStorage key: %s:',
            'navigator-test-key',
            { invalid: 'structure' }
        );

        consoleSpy.mockRestore();
    });

    it('should pass validation with valid data structure', () => {
        const validData = { name: 'John', age: 30 };
        localStorageMock.setItem(
            'navigator-test-key',
            JSON.stringify(validData)
        );

        const validator = (
            data: unknown
        ): data is { name: string; age: number } => {
            return (
                typeof data === 'object' &&
                data !== null &&
                'name' in data &&
                'age' in data &&
                typeof (data as Record<string, unknown>).name === 'string' &&
                typeof (data as Record<string, unknown>).age === 'number'
            );
        };

        const { result } = renderHook(() =>
            useLocalStorage('test-key', { name: 'default', age: 0 }, validator)
        );

        expect(result.current[0]).toEqual(validData);
        expect(localStorageMock.removeItem).not.toHaveBeenCalled();
    });

    it('should handle localStorage setItem errors gracefully', () => {
        const { result } = renderHook(() =>
            useLocalStorage('test-key', 'default-value')
        );

        // Mock setItem to throw an error
        const originalSetItem = localStorageMock.setItem;
        localStorageMock.setItem = jest.fn(() => {
            throw new Error('Storage quota exceeded');
        });

        const consoleSpy = jest
            .spyOn(console, 'warn')
            .mockImplementation(() => {});

        act(() => {
            result.current[1]('new-value');
        });

        // Should still update the state even if localStorage fails
        expect(result.current[0]).toBe('new-value');
        expect(consoleSpy).toHaveBeenCalledWith(
            'Error setting localStorage key:',
            'navigator-test-key',
            expect.any(Error)
        );

        // Restore original setItem and clean up
        localStorageMock.setItem = originalSetItem;
        consoleSpy.mockRestore();
    });

    it('should prefix localStorage keys correctly', () => {
        const { result } = renderHook(() => useLocalStorage('my-key', 'value'));

        act(() => {
            result.current[1]('test');
        });

        expect(localStorageMock.setItem).toHaveBeenCalledWith(
            'navigator-my-key',
            JSON.stringify('test')
        );
    });
});
