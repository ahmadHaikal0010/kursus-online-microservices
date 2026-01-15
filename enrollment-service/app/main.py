from fastapi import FastAPI
from .database import Base, engine
from .routers import enrollment

Base.metadata.create_all(bind=engine)

app = FastAPI(title="Enrollment Service")

app.include_router(enrollment.router)

@app.get("/health")
def health():
    return {"status": "enrollment service running"}
