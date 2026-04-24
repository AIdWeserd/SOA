FROM python:3.12-slim
WORKDIR /app
COPY movie-service/requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
COPY schema/ ./schema/
COPY movie-service/ .
CMD ["python", "main.py"]
