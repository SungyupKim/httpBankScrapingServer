package main

type DataObject interface {
	toString() string
}

type Service interface {
	execute(input DataObject) DataObject
}
