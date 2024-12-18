package controller

var (
	//kullanıcı durumu : Başarılı Mesajlar
	successMessage       = "işlem başarılı oldu"
	successMessageDelete = "silme işlemi başarılı bir şekilde gerçekleşti"

	//kullanıcı durumu : Başarısız Mesajlar
	errorMessage       = ""
	errorMessagePostID = "Beğenme işlemi sırasında hata oluştu. Lütfen daha sonra tekrar deneyiniz."

	errorMessageFindDB   = "post veri tabanında bulunamadı"
	errorMessageUid      = "kullanıcı kimliği saptanamadı"
	errorMessageDelete   = "veri silinirken hata oluştu"
	errorMessageForbiden = "bu işlem için yetkiniz bulunmamktadır"
	errorMessageLoggedIn = "lütfen giriş yapın"

	//server error
	errorrMessageInternalServer = "işlem başarısız oldu"

	//token
	errorMessageTokenNotFound = "token bulunamadı."

	//same user alredy
	errorMessageAlredyUser = "bu kullanıcı adı mevcut"
)
