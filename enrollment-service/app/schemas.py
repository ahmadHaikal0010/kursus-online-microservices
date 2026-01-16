from pydantic import BaseModel
from datetime import datetime
from uuid import UUID

class EnrollmentCreate(BaseModel):
    user_id: str
    course_id: str
    
class EnrollmentResponse(BaseModel):
    id: UUID
    user_id: str
    course_id: str
    enrolled_at: datetime
    
    class Config:
        from_attributes = True