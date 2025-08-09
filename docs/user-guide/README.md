# User Guide

This guide helps you get started with Navigator and covers common usage patterns.

## Contents

- [Installation](installation.md) - How to install Navigator
- [Getting Started](getting-started.md) - Quick start guide
- [CLI Reference](cli-reference.md) - Complete navctl command reference

## Overview

Navigator is a service-focused analysis tool for Kubernetes and Istio that provides service discovery and proxy configuration analysis. It consists of three main components:

- **navctl** - CLI tool for local development and orchestration
- **manager** - Central coordination point for edge connections
- **edge** - Connects to Kubernetes clusters and streams state

## Quick Start

1. [Install Navigator](installation.md)
2. Run `navctl local` to start all services locally
3. Open your browser to view the service registry UI

For detailed instructions, see the [Getting Started](getting-started.md) guide.