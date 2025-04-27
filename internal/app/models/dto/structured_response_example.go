package dto

// This file contains examples of how to use the structured response DTOs

/*
Example of a flat response (old format):
{
  "success": true,
  "message": "User profile retrieved successfully",
  "data": {
    "id": 1,
    "email": "user@school.edu.tr",
    "firstName": "John",
    "lastName": "Doe",
    "roleType": "STUDENT",
    "profilePhotoFileId": 123,
    "identifier": "12345678",
    "graduationYear": 2025,
    "departmentId": 1,
    "departmentName": "Computer Engineering",
    "facultyId": 1,
    "facultyName": "Engineering Faculty"
  },
  "timestamp": "2024-07-25T14:30:00Z"
}
*/

/*
Example of a structured response (new format):
{
  "success": true,
  "message": "User profile retrieved successfully",
  "data": {
    "id": 1,
    "email": "user@school.edu.tr",
    "firstName": "John",
    "lastName": "Doe",
    "roleType": "STUDENT",
    "profile": {
      "profilePhoto": {
        "id": 123
      },
      "identifier": "12345678",
      "graduationYear": 2025
    },
    "department": {
      "id": 1,
      "name": "Computer Engineering",
      "facultyId": 1
    },
    "faculty": {
      "id": 1,
      "name": "Engineering Faculty"
    }
  },
  "timestamp": "2024-07-25T14:30:00Z"
}
*/

// Example usage in a controller/handler:
/*
func GetUserProfile(c *gin.Context) {
    // Get user profile from service
    userProfile, err := userService.GetUserProfile(userID)
    if err != nil {
        // Handle error
        return
    }

    // Convert to structured format
    structuredData := dto.NewStructuredUserResponse(userProfile)

    // Return response
    response := dto.NewStructuredResponse(structuredData, "User profile retrieved successfully")
    c.JSON(http.StatusOK, response)
}
*/

// Example with pagination:
/*
{
  "success": true,
  "message": "User list retrieved successfully",
  "data": {
    "items": [
      {
        "id": 1,
        "email": "user1@school.edu.tr",
        "firstName": "John",
        "lastName": "Doe",
        "roleType": "STUDENT",
        "profile": { ... },
        "department": { ... },
        "faculty": { ... }
      },
      {
        "id": 2,
        "email": "user2@school.edu.tr",
        "firstName": "Jane",
        "lastName": "Smith",
        "roleType": "INSTRUCTOR",
        "profile": { ... },
        "department": { ... },
        "faculty": { ... }
      }
    ],
    "pagination": {
      "currentPage": 0,
      "totalPages": 5,
      "pageSize": 10,
      "totalItems": 48
    }
  },
  "timestamp": "2024-07-25T14:30:00Z"
}
*/
