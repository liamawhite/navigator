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

import { renderHook, act } from '@testing-library/react';
import { useCollapsibleSections } from './useCollapsibleSections';

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

describe('useCollapsibleSections', () => {
    const defaultGroups = {
        section1: false,
        section2: true,
        section3: false,
    };

    beforeEach(() => {
        localStorageMock.clear();
        jest.clearAllMocks();
    });

    it('should initialize with default groups', () => {
        const { result } = renderHook(() =>
            useCollapsibleSections('test-sections', defaultGroups)
        );

        expect(result.current.collapsedGroups).toEqual(defaultGroups);
        expect(result.current.isGroupCollapsed('section1')).toBe(false);
        expect(result.current.isGroupCollapsed('section2')).toBe(true);
        expect(result.current.isGroupCollapsed('section3')).toBe(false);
    });

    it('should toggle group collapse state', () => {
        const { result } = renderHook(() =>
            useCollapsibleSections('test-sections', defaultGroups)
        );

        // Toggle section1 from false to true
        act(() => {
            result.current.toggleGroupCollapse('section1');
        });

        expect(result.current.isGroupCollapsed('section1')).toBe(true);
        expect(result.current.collapsedGroups.section1).toBe(true);

        // Toggle section1 back to false
        act(() => {
            result.current.toggleGroupCollapse('section1');
        });

        expect(result.current.isGroupCollapsed('section1')).toBe(false);
        expect(result.current.collapsedGroups.section1).toBe(false);
    });

    it('should toggle section2 from true to false', () => {
        const { result } = renderHook(() =>
            useCollapsibleSections('test-sections', defaultGroups)
        );

        // section2 starts as true
        expect(result.current.isGroupCollapsed('section2')).toBe(true);

        // Toggle section2 to false
        act(() => {
            result.current.toggleGroupCollapse('section2');
        });

        expect(result.current.isGroupCollapsed('section2')).toBe(false);
        expect(result.current.collapsedGroups.section2).toBe(false);
    });

    it('should persist state to localStorage', () => {
        const { result } = renderHook(() =>
            useCollapsibleSections('test-sections', defaultGroups)
        );

        act(() => {
            result.current.toggleGroupCollapse('section1');
        });

        expect(localStorageMock.setItem).toHaveBeenCalledWith(
            'navigator-test-sections',
            JSON.stringify({
                section1: true,
                section2: true,
                section3: false,
            })
        );
    });

    it('should load saved state from localStorage', () => {
        const savedState = {
            section1: true,
            section2: false,
            section3: true,
        };
        localStorageMock.setItem(
            'navigator-test-sections',
            JSON.stringify(savedState)
        );

        const { result } = renderHook(() =>
            useCollapsibleSections('test-sections', defaultGroups)
        );

        expect(result.current.collapsedGroups).toEqual(savedState);
        expect(result.current.isGroupCollapsed('section1')).toBe(true);
        expect(result.current.isGroupCollapsed('section2')).toBe(false);
        expect(result.current.isGroupCollapsed('section3')).toBe(true);
    });

    it('should handle invalid localStorage data', () => {
        // Set invalid data
        localStorageMock.setItem('navigator-test-sections', 'invalid-json{');

        const consoleSpy = jest
            .spyOn(console, 'warn')
            .mockImplementation(() => {});

        const { result } = renderHook(() =>
            useCollapsibleSections('test-sections', defaultGroups)
        );

        // Should fall back to default groups
        expect(result.current.collapsedGroups).toEqual(defaultGroups);
        expect(localStorageMock.removeItem).toHaveBeenCalledWith(
            'navigator-test-sections'
        );

        consoleSpy.mockRestore();
    });

    it('should validate data structure', () => {
        // Set data with wrong structure
        localStorageMock.setItem(
            'navigator-test-sections',
            JSON.stringify({ section1: 'invalid' })
        );

        const consoleSpy = jest
            .spyOn(console, 'warn')
            .mockImplementation(() => {});

        const { result } = renderHook(() =>
            useCollapsibleSections('test-sections', defaultGroups)
        );

        // Should fall back to default groups
        expect(result.current.collapsedGroups).toEqual(defaultGroups);
        expect(localStorageMock.removeItem).toHaveBeenCalledWith(
            'navigator-test-sections'
        );

        consoleSpy.mockRestore();
    });

    it('should return false for unknown group keys', () => {
        const { result } = renderHook(() =>
            useCollapsibleSections('test-sections', defaultGroups)
        );

        // @ts-expect-error - Testing runtime behavior
        expect(result.current.isGroupCollapsed('unknown-section')).toBe(false);
    });

    it('should handle different storage keys independently', () => {
        const { result: result1 } = renderHook(() =>
            useCollapsibleSections('sections-1', { group1: false })
        );
        const { result: result2 } = renderHook(() =>
            useCollapsibleSections('sections-2', { group1: true })
        );

        expect(result1.current.isGroupCollapsed('group1')).toBe(false);
        expect(result2.current.isGroupCollapsed('group1')).toBe(true);

        act(() => {
            result1.current.toggleGroupCollapse('group1');
        });

        expect(result1.current.isGroupCollapsed('group1')).toBe(true);
        expect(result2.current.isGroupCollapsed('group1')).toBe(true); // Should remain unchanged
    });
});
