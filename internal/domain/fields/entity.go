package fields

type EntityField string

const (
	EntityID         EntityField = "id"
	EntityName       EntityField = "name"
	EntityEntityType EntityField = "entity_type"
	EntityCreatorID  EntityField = "creator_id"
	EntityParentID   EntityField = "parent_id"
	EntityParentType EntityField = "parent_type"
	EntityCreatedAt  EntityField = "created_at"
	EntityUpdatedAt  EntityField = "updated_at"
	EntityIsDeleted  EntityField = "is_deleted"
)

func (f EntityField) APIName() string {
	return string(f)
}
func (f EntityField) FirestoreName() string {
	return string(f)
}

func (f EntityField) DomainName() string {
	switch f {
	case EntityID:
		return "ID"
	case EntityName:
		return "Name"
	case EntityEntityType:
		return "EntityType"
	case EntityCreatorID:
		return "CreatorID"
	case EntityParentID:
		return "ParentID"
	case EntityParentType:
		return "ParentType"
	case EntityCreatedAt:
		return "CreatedAt"
	case EntityUpdatedAt:
		return "UpdatedAt"
	case EntityIsDeleted:
		return "IsDeleted"
	default:
		return ""
	}
}

func (f EntityField) IsValid() bool {
	switch f {
	case EntityID, EntityName, EntityEntityType, EntityCreatorID, EntityParentID, EntityParentType, EntityCreatedAt, EntityUpdatedAt, EntityIsDeleted:
		return true
	default:
		return false
	}
}

var EntityFields = []EntityField{
	EntityID, EntityName, EntityEntityType, EntityCreatorID, EntityParentID, EntityParentType, EntityCreatedAt, EntityUpdatedAt, EntityIsDeleted,
}
