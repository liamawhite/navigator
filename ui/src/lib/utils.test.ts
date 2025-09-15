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

// Jest globals are available globally
import { cn, formatLastUpdated } from './utils';

describe('cn', () => {
    it('should merge class names correctly', () => {
        expect(cn('foo', 'bar')).toBe('foo bar');
    });

    it('should handle conditional classes', () => {
        const shouldShow = false;
        expect(cn('foo', shouldShow && 'bar', 'baz')).toBe('foo baz');
    });

    it('should merge conflicting tailwind classes', () => {
        expect(cn('px-2', 'px-4')).toBe('px-4');
    });

    it('should handle empty input', () => {
        expect(cn()).toBe('');
    });

    it('should handle complex class combinations', () => {
        const isHidden = false;
        const result = cn(
            'px-2 py-1',
            'hover:bg-blue-500',
            isHidden && 'hidden',
            'text-white'
        );
        expect(result).toBe('px-2 py-1 hover:bg-blue-500 text-white');
    });
});

describe('formatLastUpdated', () => {
    const fixedTime = new Date('2024-01-01T12:00:00Z');

    beforeAll(() => {
        jest.useFakeTimers();
        jest.setSystemTime(fixedTime);
    });

    afterAll(() => {
        jest.useRealTimers();
    });

    it('should return "Never" for null date', () => {
        expect(formatLastUpdated(null)).toBe('Never');
    });

    it('should format seconds correctly', () => {
        const date = new Date('2024-01-01T11:59:30Z'); // 30 seconds ago
        expect(formatLastUpdated(date)).toBe('30s ago');
    });

    it('should format minutes correctly', () => {
        const date = new Date('2024-01-01T11:45:00Z'); // 15 minutes ago
        expect(formatLastUpdated(date)).toBe('15m ago');
    });

    it('should format hours correctly', () => {
        const date = new Date('2024-01-01T09:00:00Z'); // 3 hours ago
        expect(formatLastUpdated(date)).toBe('3h ago');
    });

    it('should handle edge cases', () => {
        // Exactly 1 minute ago
        const oneMinuteAgo = new Date('2024-01-01T11:59:00Z');
        expect(formatLastUpdated(oneMinuteAgo)).toBe('1m ago');

        // Exactly 1 hour ago
        const oneHourAgo = new Date('2024-01-01T11:00:00Z');
        expect(formatLastUpdated(oneHourAgo)).toBe('1h ago');

        // Less than 1 second ago
        const justNow = new Date('2024-01-01T12:00:00Z');
        expect(formatLastUpdated(justNow)).toBe('0s ago');
    });
});
