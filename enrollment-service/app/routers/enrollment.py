from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List
import requests
import os

from ..database import SessionLocal
from ..models import Enrollment
from ..schemas import EnrollmentCreate, EnrollmentResponse

router = APIRouter(prefix="/api/enrollments", tags=["Enrollments"])

COURSE_SERVICE_URL = os.getenv("COURSE_SERVICE_URL")

def get_db():
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()
        
@router.post("", response_model=EnrollmentResponse)
def create_enrollment(data: EnrollmentCreate, db: Session = Depends(get_db)):

    # 1. VALIDASI COURSE (API CALL)
    course_res = requests.get(f"{COURSE_SERVICE_URL}/api/courses/{data.course_id}")
    if course_res.status_code != 200:
        raise HTTPException(status_code=400, detail="Course not found")

    # 2. SIMPAN ENROLLMENT
    enrollment = Enrollment(
        user_id=data.user_id,
        course_id=data.course_id
    )
    db.add(enrollment)
    db.commit()
    db.refresh(enrollment)

    return enrollment

@router.get("/user/{user_id}", response_model=List[EnrollmentResponse])
def get_enrollments_by_user(user_id: str, db: Session = Depends(get_db)):
    enrollments = (
        db.query(Enrollment)
        .filter(Enrollment.user_id == user_id)
        .all()
    )
    return enrollments