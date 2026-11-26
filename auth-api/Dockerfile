FROM maven:3.9.9-eclipse-temurin-21 AS builder

WORKDIR /build
COPY pom.xml .
RUN mvn dependency:go-offline -B

COPY src ./src
RUN mvn clean package -DskipTests -B

FROM eclipse-temurin:21-jre

WORKDIR /app
RUN mkdir -p /data
ENV DB_PATH=/data/auth.db

COPY --from=builder /build/target/auth-api-1.0.0.jar app.jar

EXPOSE 8001

ENTRYPOINT ["java", "-jar", "/app/app.jar"]
