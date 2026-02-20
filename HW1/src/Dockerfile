FROM python:3.11-slim

WORKDIR /user_service

COPY user_service/ ./user_service/

EXPOSE 8000

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8000/health || exit 1

CMD ["python", "user_service/main.py"]