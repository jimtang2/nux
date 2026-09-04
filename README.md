# Enhanced User Experience -- nux

## 1. What This Is

In a nutshell, nux is a framework for creating and applying rules on UI, UX, and API interactions based on user profile context.

This package contains a set of tools that solve various problems in deploying effective personalized customer strategies through UI, UX, and APIs.

## 2. What This Is NOT

This is not a telemetry library. I recommend [OpenTelemetry](https://opentelemetry.io/) for normalized datasets.


## 3. Overview

Users interact with online businesses via UI (web, mobile, desktop). In order to provide personalized experience, services 

### Feedback Loop



## Architecture

### 

#### Kafka

- Stores users_events

### Rules Engine 

#### Postgresql

- Stores rules, profile definitions, versions
- Stores users_profiles

#### Redis

- Stores users profiles

### API 

- Returns users profiles