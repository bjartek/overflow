 - OverflowState.Account is gone use AccountE and handle the error: (might consider adding this back again and deprecating it)
 - Script with inline -> InlineScript
 - the type of the state changed from Overflow -> OverflowState
 - parsing of events into interface{}/json is more typesafe so it changes from not only beeing strings!
 - remove discordgo
 - changed the folder structure so you do not have to double import overflow and make it build better on godoc

 TODO:
  - document what not to do when composing.

