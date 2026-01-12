from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field
from database import reviews_collection
from fastapi import Body
from bson import ObjectId

app = FastAPI()


class Review(BaseModel):
    product_id: int = Field(..., example=1)
    review: str = Field(..., example="Barang bagus, pengiriman cepat")
    rating: int = Field(..., ge=1, le=5, example=5)


@app.post("/reviews")
def create_review(review: Review):
    review_dict = review.model_dump()
    result = reviews_collection.insert_one(review_dict)
    review_dict["_id"] = str(result.inserted_id)

    return {
        "success": True,
        "data": review_dict
    }
    
    




@app.get("/reviews")
def get_reviews():
    try:
        reviews = []
        for r in reviews_collection.find():
            r["_id"] = str(r["_id"])  # FIX ObjectId
            reviews.append(r)

        return {
            "success": True,
            "data": reviews
        }

    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@app.get("/reviews/product/{product_id}")
def get_reviews_by_product(product_id: int):
    reviews = []
    for r in reviews_collection.find({"product_id": product_id}):
        r["_id"] = str(r["_id"])
        reviews.append(r)

    return {
        "success": True,
        "data": reviews
    }

    
    
@app.delete("/reviews/{review_id}")
def delete_review(review_id: str):
    result = reviews_collection.delete_one(
        {"_id": ObjectId(review_id)}
    )

    if result.deleted_count == 0:
        raise HTTPException(status_code=404, detail="Review tidak ditemukan")

    return {"success": True}


@app.put("/reviews/{review_id}")
def update_review(review_id: str, review: Review = Body(...)):
    review_dict = review.model_dump()
    
    result = reviews_collection.update_one(
        {"_id": ObjectId(review_id)},
        {"$set": review_dict}
    )
    
    if result.matched_count == 0:
        raise HTTPException(status_code=404, detail="Review tidak ditemukan")
    
    review_dict["_id"] = review_id
    return {
        "success": True,
        "data": review_dict
    }
