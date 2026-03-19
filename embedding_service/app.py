from flask import Flask, request, jsonify
from sentence_transformers import SentenceTransformer

app = Flask(__name__)
model = SentenceTransformer('intfloat/multilingual-e5-large')


@app.route('/embed', methods=['POST'])
def embed():
    data = request.json or {}
    text = data.get('text', '')
    embed_type = data.get('type', 'query')
    prefix = "passage: " if embed_type == "product" else "query: "
    embedding = model.encode(prefix + text)
    return jsonify({'embedding': embedding.tolist()})


@app.route('/health', methods=['GET'])
def health():
    return jsonify({'status': 'ok'})


if __name__ == '__main__':
    app.run(host='0.0.0.0', port=8000)
