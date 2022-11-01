package overflow

type MetadataViews_HTTPFile struct {
	Url string
}

type MetadataViews_IPFSFile struct {
	Cid  string
	Path *string
}

type MetadataViews_Display_IPFS struct {
	Name        string
	Description string
	Thumbnail   MetadataViews_IPFSFile
}
type MetadataViews_Display_Http struct {
	Name        string
	Description string
	Thumbnail   MetadataViews_HTTPFile
}

type MetadataViews_Edition struct {
	Name   *string
	Number uint64
	Max    *uint64
}

type MetadataViews_Editions struct {
	Editions []MetadataViews_Edition `cadence:"infoList"`
}

type MetadataViews_Serial struct {
	Number uint64
}

type MetadataViews_Media_IPFS struct {
	File      MetadataViews_IPFSFile
	MediaType string `cadence:"mediaType"`
}

type MetadataViews_Media_HTTP struct {
	File      MetadataViews_HTTPFile
	MediaType string `cadence:"mediaType"`
}
type MetadtaViews_Licensce struct {
	Spdx string `cadence:"spdxIdentifier"`
}

type MetadataViews_ExternalURL struct {
	Url string
}

type MetadataViews_Rarity struct {
	Score       *string
	Max         *uint64
	Description *string
}

type MetadataViews_Trait struct {
	Name        string
	Value       interface{}
	DisplayType string `cadence:"displayType"`
	Rarity      *MetadataViews_Rarity
}

type MetadataViews_Traits struct {
	Traits []MetadataViews_Trait
}
