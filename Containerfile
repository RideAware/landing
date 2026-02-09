# Build stage
FROM mcr.microsoft.com/dotnet/sdk:8.0-alpine AS builder

WORKDIR /app

# Copy csproj and restore dependencies
COPY landing.csproj .
RUN dotnet restore

# Copy everything and publish
COPY . .
RUN dotnet publish -c Release -o /app/publish --no-restore

# Runtime stage
FROM mcr.microsoft.com/dotnet/aspnet:8.0-alpine

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/publish .

EXPOSE 5000

ENV ASPNETCORE_ENVIRONMENT=Production
ENV DOTNET_RUNNING_IN_CONTAINER=true

CMD ["dotnet", "landing.dll"]
