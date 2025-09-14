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

import { describe, it, expect, jest } from '@jest/globals';
import { render, screen, fireEvent } from '@testing-library/react';
import { ServiceCard } from './ServiceCard';
import type { v1alpha1Service } from '../../types/generated/openapi-service_registry/models/v1alpha1Service';

describe('ServiceCard', () => {
    const mockService: v1alpha1Service = {
        name: 'test-service',
        namespace: 'default',
        instances: [
            {
                instanceId: 'instance-1',
                envoyPresent: true,
            },
            {
                instanceId: 'instance-2',
                envoyPresent: false,
            },
            {
                instanceId: 'instance-3',
                envoyPresent: true,
            },
        ],
    };

    it('should render service name and namespace', () => {
        render(<ServiceCard service={mockService} />);

        expect(screen.getByText('test-service')).toBeTruthy();
        expect(screen.getByText('default')).toBeTruthy();
    });

    it('should display correct proxy count', () => {
        render(<ServiceCard service={mockService} />);

        // Check for instance count and proxy info
        expect(screen.getByText(/3.*instances?/)).toBeTruthy();
        expect(screen.getByText(/2.*with Envoy/)).toBeTruthy();
    });

    it('should call onClick handler when clicked', () => {
        const mockOnClick = jest.fn();
        render(<ServiceCard service={mockService} onClick={mockOnClick} />);

        const card = screen.getByText('test-service').closest('div');
        if (card) {
            fireEvent.click(card);
        }

        expect(mockOnClick).toHaveBeenCalledTimes(1);
    });
});
