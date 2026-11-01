package config

func MongoURI() string          { return GetEnv("MONGO_URI", "mongodb://127.0.0.1:27017") }
func MongoDatabaseName() string { return GetEnv("MONGO_DATABASE", "pertemuan11_mongo") }
