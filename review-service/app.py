from flask import Flask, jsonify, request

app = Flask(__name__)

# dummy data for reviews
reviews = [
    {"id": 1, "product_id": 101, "review": "Great product!", "rating": 5},
    {"id": 2, "product_id": 102, "review": "Not bad", "rating": 3},
    {"id": 3, "product_id": 101, "review": "Could be better", "rating": 2},
]

# GET /reviews -> ambil semua reviews
@app.route('/reviews', methods=['GET'])
def get_all_reviews():
    return jsonify(reviews)

# GET /reviews/all -> ambil semua reviews
@app.route('/reviews/all', methods=['GET'])
def get_reviews():
    return jsonify(reviews)

# GET /reviews/<review_id> -> review berdasarkan review_id
@app.route('/reviews/<int:review_id>', methods=['GET'])
def get_review_by_id(review_id):
    review = next((r for r in reviews if r["id"] == review_id), None)
    if review is None:
        return jsonify({"message": "Review not found"}), 404
    return jsonify(review)

# DELETE /reviews/<review_id> -> hapus review
@app.route('/reviews/<int:review_id>', methods=['DELETE'])
def delete_review(review_id):
    global reviews

    review = next((r for r in reviews if r["id"] == review_id), None)
    if review is None:
        return jsonify({"message": "Review not found"}), 404

    reviews = [r for r in reviews if r["id"] != review_id]

    return jsonify({"message": "Review deleted"}), 200


# GET /reviews/product/<product_id> -> reviews berdasarkan product_id
@app.route('/reviews/product/<int:product_id>', methods=['GET'])
def get_reviews_by_product(product_id):
    filtered_reviews = [r for r in reviews if r["product_id"] == product_id]
    return jsonify(filtered_reviews)

# POST /reviews -> tambah review baru
@app.route('/reviews', methods=['POST'])
def create_review():
    data = request.get_json() or {}

    required_fields = ["product_id", "review", "rating"]
    missing = [field for field in required_fields if field not in data]
    
    if missing:
        return jsonify({"message": f"Missing fields: {', '.join(missing)}"}), 400

    new_id = len(reviews) + 1
    new_review = {
        "id": new_id,
        "product_id": int(data["product_id"]),
        "review": data["review"],
        "rating": int(data["rating"])
    }

    reviews.append(new_review)

    return jsonify({"message": "Review created", "data": new_review}), 201

if __name__ == "__main__":
    app.run(host="0.0.0.0", port    =5002, debug=True)
