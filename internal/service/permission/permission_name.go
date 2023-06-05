package permission

const (
	AdminAccess               = "admin.access"
	QuestionAdd               = "question.add"
	QuestionEdit              = "question.edit"
	QuestionEditWithoutReview = "question.edit_without_review"
	QuestionDelete            = "question.delete"
	QuestionClose             = "question.close"
	QuestionReopen            = "question.reopen"
	QuestionVoteUp            = "question.vote_up"
	QuestionVoteDown          = "question.vote_down"
	QuestionPin               = "question.pin"   //Top  the question
	QuestionUnPin             = "question.unpin" //untop the question
	QuestionHide              = "question.hide"  //hide  the question
	QuestionShow              = "question.show"  //show the question
	AnswerAdd                 = "answer.add"
	AnswerEdit                = "answer.edit"
	AnswerEditWithoutReview   = "answer.edit_without_review"
	AnswerDelete              = "answer.delete"
	AnswerAccept              = "answer.accept"
	AnswerVoteUp              = "answer.vote_up"
	AnswerVoteDown            = "answer.vote_down"
	CommentAdd                = "comment.add"
	CommentEdit               = "comment.edit"
	CommentDelete             = "comment.delete"
	CommentVoteUp             = "comment.vote_up"
	CommentVoteDown           = "comment.vote_down"
	ReportAdd                 = "report.add"
	TagAdd                    = "tag.add"
	TagEdit                   = "tag.edit"
	TagEditSlugName           = "tag.edit_slug_name"
	TagEditWithoutReview      = "tag.edit_without_review"
	TagDelete                 = "tag.delete"
	TagSynonym                = "tag.synonym"
	LinkUrlLimit              = "link.url_limit"
	VoteDetail                = "vote.detail"
	AnswerAudit               = "answer.audit"
	QuestionAudit             = "question.audit"
	TagAudit                  = "tag.audit"
	TagUseReservedTag         = "tag.use_reserved_tag"
)

const (
	reportActionName = "action.report"
	editActionName   = "action.edit"
	deleteActionName = "action.delete"
	closeActionName  = "action.close"
	reopenActionName = "action.reopen"
	pinActionName    = "action.pin"
	unpinActionName  = "action.unpin"
	hideActionName   = "action.hide"
	showActionName   = "action.show"
)
