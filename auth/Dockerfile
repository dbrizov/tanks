# syntax=docker/dockerfile:1

# --- Build stage: produce the Spring Boot fat JAR ---
FROM eclipse-temurin:25-jdk AS build
WORKDIR /src

# Warm the Gradle dependency cache as a separate layer before copying sources.
COPY gradlew settings.gradle build.gradle ./
COPY gradle ./gradle
RUN chmod +x ./gradlew && ./gradlew --no-daemon dependencies || true

COPY . .
# `build` runs the test suite and produces the fat bootJar (fail-fast on
# a broken build before we ship an image).
RUN ./gradlew --no-daemon clean build

# --- Runtime stage: JRE only ---
FROM eclipse-temurin:25-jre
WORKDIR /app
COPY --from=build /src/build/libs/*.jar app.jar
EXPOSE 8081
# Prod profile expects DB_URL, DB_USERNAME, DB_PASSWORD, JWT_SECRET (and
# optionally APP_CORS_ALLOWED_ORIGINS) from the environment.
ENV SPRING_PROFILES_ACTIVE=prod
ENTRYPOINT ["java", "-jar", "app.jar"]
