FROM eclipse-temurin:17-jdk-alpine

WORKDIR /app

COPY target/review-service-1.0.jar app.jar

EXPOSE 5002

CMD ["java", "-jar", "app.jar"]
