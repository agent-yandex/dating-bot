package states

type State string

const (
	StateDefault             State = "default"
	StateEditName            State = "editing_name"
	StateEditAge             State = "editing_age"
	StateEditBio             State = "editing_bio"
	StateEditGender          State = "editing_gender"
	StateEditCity            State = "editing_city"
	StateEditPrefMinage      State = "edit_pref_min_age"
	StateEditPrefMaxAge      State = "edit_pref_max_age"
	StateEditPrefGender      State = "edit_pref_gender"
	StateEditPrefMaxDistance State = "edit_pref_max_distance"
	StateSearching           State = "searching"
	StateViewLikes           State = "view_likes"
	StateEditPhoto           State = "editing_photo"
)
