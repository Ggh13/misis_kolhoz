FROM python:3.11

WORKDIR /app

# CPU-only PyTorch (smaller)
RUN pip install --no-cache-dir torch --index-url https://download.pytorch.org/whl/cpu

# ML dependencies
RUN pip install --no-cache-dir \
    fastapi uvicorn python-dotenv openai deep-translator \
    transformers sentencepiece protobuf pandas openpyxl

# Pre-download models into the image (no download on startup)
ARG HF_TOKEN
ENV HF_TOKEN=${HF_TOKEN}
COPY download_models.py /app/download_models.py
RUN python3 /app/download_models.py

COPY . .

EXPOSE 8000

CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]