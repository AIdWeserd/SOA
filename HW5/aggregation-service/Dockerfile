FROM python:3.12-slim
WORKDIR /app
COPY aggregation-service/requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
COPY aggregation-service/ .
CMD ["python", "main.py"]
