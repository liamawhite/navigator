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

import { useCallback } from 'react';
import { useLocalStorage } from './useLocalStorage';

/**
 * Validator function for collapsible sections data structure
 */
const validateCollapsibleData = <T extends string>(
    data: unknown
): data is Record<T, boolean> => {
    return (
        typeof data === 'object' &&
        data !== null &&
        Object.values(data).every((v) => typeof v === 'boolean')
    );
};

/**
 * Custom hook for managing collapsible sections with localStorage persistence
 * @param storageKey - The localStorage key to use for persistence
 * @param defaultGroups - Default collapse state for each section
 * @returns Object containing collapsed state and toggle function
 */
export function useCollapsibleSections<T extends string>(
    storageKey: string,
    defaultGroups: Record<T, boolean>
) {
    const [collapsedGroups, setCollapsedGroups] = useLocalStorage<
        Record<T, boolean>
    >(storageKey, defaultGroups, validateCollapsibleData<T>);

    const toggleGroupCollapse = useCallback(
        (groupKey: T) => {
            setCollapsedGroups((prev) => ({
                ...prev,
                [groupKey]: !prev[groupKey],
            }));
        },
        [setCollapsedGroups]
    );

    return {
        collapsedGroups,
        toggleGroupCollapse,
        isGroupCollapsed: (groupKey: T) => collapsedGroups[groupKey] ?? false,
    };
}
