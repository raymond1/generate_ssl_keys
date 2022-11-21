package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

var rootAuthorityPrivateKeyFilename = "root.pem"
var intermediateAuthorityPrivateKeyFilename = "intermediate.pem"
var serverPrivateKeyFilename = "server.pem"

var outputDirectory = "output"

var templatesDirectory = "templates"

var rootAuthorityCSRConfigFilename = "root_csr.conf"
var rootAuthorityConfigurationFilename = "root_ca.conf"

var rootAuthorityDirectory = "" //Will be set in makeDirectories
var rootAuthorityDirectoryName = "root_authority"
var rootAuthorityDatabaseFilename = "root_database.txt"
var rootAuthoritySerialNumberFilename = "root_serial_number.txt"

var intermediateAuthorityCSRConfigFilename = "intermediate_csr.conf"
var intermediateAuthorityConfigurationFilename = "intermediate_ca.conf"

var intermediateAuthorityDirectory = "" //Will be set in makeDirectories
var intermediateAuthorityDirectoryName = "intermediate_directory"
var intermediateAuthorityDatabaseFilename = "intermediate_database.txt"
var intermediateAuthoritySerialNumberFilename = "intermediate_serial_number.txt"

//Takes in a string and runs the command in a shell
func runCommand(command string) error {
	executableCommand := convertStringIntoExecCommand(command)
	return executableCommand.Run()
}

//Takes in a string and produces Cmd object that can be run
func convertStringIntoExecCommand(command string) *exec.Cmd {
	arguments := strings.Split(command, " ")
	return exec.Command(arguments[0], arguments[1:]...)
}

//Incorrect
func signCertificate(signerPrivateKey string, certificateSigningRequest string, outputCertificate string) {
	//openssl ca -selfsign -keyfile root.pem -config root_ca.conf -out root.crt -in root.csr -outdir root_certificates -verbose -batch
	command := ""
	executableCommand := convertStringIntoExecCommand(command)
	executableCommand.Run()
}

//Overview: Takes in the root private key along with some other required information and generates a certificate for it.
//
//The specific parameters required are:
//rootPrivateKey: a string to the file containing the root private key
//certificateSigningRequest: a string path to the certificate signing request file. (The certificate signing request contains the information that will be signed into the certificate.)
//rootAuthorityConfiguration: a string path to the root authority configuration file
//outputCertificate: a string path specifying the output file storing the generated certificate
//domainName: For this program, the domain name is used as the folder name where output files are stored.
func generateSelfSignedCertificate(rootPrivateKey string, certificateSigningRequest string, rootAuthorityConfiguration string, domainName string, outputDirectory string, outputCertificate string) {

	command := fmt.Sprintf("openssl ca -selfsign -keyfile %s -config %s -out %s -in %s -outdir %s -verbose -batch",
		rootPrivateKey, rootAuthorityConfiguration, outputCertificate, certificateSigningRequest, outputDirectory)

	fmt.Println("Inside generateSelfSignedCertificate", command)
	err := runCommand(command)
	if err != nil {
		fmt.Println("Error during generation of self-signed certificate. Command was: " + command)
		fmt.Println(err)
		os.Exit(0)
	}
}

//privateKey is a string specifying the filepath of a private key file
//configurationFilepath is a string specifying the filepath of a configuration file for signing the certificate
//outputCertificate is a string specifying the filepath of the certificate that will be generated by running this generateCertificateSigningRequest
func generateCertificateSigningRequest(privateKey string, configurationFilepath string, outputCertificate string) {
	command := fmt.Sprintf("openssl req -key %s -out %s -days 398 -new -config %s", privateKey, outputCertificate, configurationFilepath)
	err := runCommand(command)
	if err != nil {
		fmt.Println("An error occurred when trying to generate the certificate signing request for key " + privateKey + " using the configuration " + configurationFilepath)
		fmt.Println("The command was:", command)
		fmt.Println(err)
	}
}

//Generates a signed certificate using the openssl ca command
func generateSignedCertificate(certificateSigningRequest, outputCertificateFilename, certificateAuthorityConfiguration, certificateAuthoritySigningKey, certificateAuthorityCertificate, outputCertificateDirectory string) {
	command := fmt.Sprintf("openssl ca -in %s -out %s -config %s -keyfile %s -cert %s -outdir %s -batch", certificateSigningRequest, outputCertificateFilename, certificateAuthorityConfiguration, certificateAuthoritySigningKey, certificateAuthorityCertificate, outputCertificateDirectory)
	err := runCommand(command)
	if err != nil {
		fmt.Println("An error occurred when trying to generate the signed certificate. The command was: ", command)
		fmt.Println(err)
	}
}

//Uses OpenSSL to generate a private key
func generatePrivateKey(filename string) error {
	command := fmt.Sprintf("openssl genpkey -outform pem -out %s -algorithm rsa", filename)
	err := runCommand(command)

	if err != nil {
		fmt.Println("An error occurred when trying to generate private key " + filename + " using OpenSSL.")
		fmt.Println(err)
	}
	return err
}

//Returns true if file or directory passed in exists
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

//Given the domain name directory, generates the directory structure needed for this program
func makeDirectories(domainNameDirectory string) {
	//1)Ensure a directory named after the domain name passed in always exists in the output directory
	if !fileExists(outputDirectory) {
		fmt.Println("Generating directory: " + outputDirectory)
		os.Mkdir(outputDirectory, 0700)
	}

	if !fileExists(domainNameDirectory) {
		fmt.Println("Generating directory: " + domainNameDirectory)
		os.Mkdir(domainNameDirectory, 0700)
	}

	rootAuthorityDirectory = domainNameDirectory + "/" + rootAuthorityDirectoryName
	if !fileExists(rootAuthorityDirectory) {
		fmt.Println("Generating directory: " + rootAuthorityDirectory)
		os.Mkdir(rootAuthorityDirectory, 0700)
	}

	intermediateAuthorityDirectory = domainNameDirectory + "/" + intermediateAuthorityDirectoryName
	if !fileExists(intermediateAuthorityDirectory) {
		fmt.Println("Generating directory: " + intermediateAuthorityDirectory)
		os.Mkdir(intermediateAuthorityDirectory, 0700)
	}
}

//Takes in an output directory and generates 3 private keys, one for the root authority, one for the intermediate authority, and one for the server hosting the domain name.
func makePrivateKeys(domainNameDirectory string) {
	//2)Create a root authority private key if it doesn't already exist. Do not replace an existing one
	//openssl genpkey -outform pem -out root.pem -algorithm rsa
	rootAuthorityPrivateKey := fmt.Sprintf("%s/%s", domainNameDirectory, rootAuthorityPrivateKeyFilename)
	if !fileExists(rootAuthorityPrivateKey) {
		fmt.Println("Generating root private key: " + rootAuthorityPrivateKey)
		err := generatePrivateKey(rootAuthorityPrivateKey)
		if err != nil {
			os.Exit(0)
		}
	}

	//3)Create an intermediate authority private key
	intermediateAuthorityPrivateKey := fmt.Sprintf("%s/%s", domainNameDirectory, intermediateAuthorityPrivateKeyFilename)
	if !fileExists(intermediateAuthorityPrivateKey) {
		fmt.Println("Generating intermediate private key: " + rootAuthorityPrivateKey)
		generatePrivateKey(intermediateAuthorityPrivateKey)
	}

	//4)Generate a server private key
	serverPrivateKey := fmt.Sprintf("%s/%s", domainNameDirectory, serverPrivateKeyFilename)
	if !fileExists(serverPrivateKey) {
		fmt.Println("Generating server private key: " + serverPrivateKey)
		generatePrivateKey(serverPrivateKey)
	}
}

//Copies the file in source to destination
func fileCopy(src, dst string) {
	fmt.Println("src:", src, "dst:", dst)
	bytesRead, err := ioutil.ReadFile(src)

	if err != nil {
		fmt.Println(err)
	}

	err = ioutil.WriteFile(dst, bytesRead, 0644)

	if err != nil {
		fmt.Println(err)
	}
}

//templateFilepath: a template openssl configuration file
//outputConfigurationFilepath: the output file that will be guaranteed to exist. One will be generated by copying
//from the templates directory if it does not already exist.
//databaseFilepath and serialNumberFilepath are the file locations of the database and serial number files used
//for the configuration file used for the openssl ca command.
func hydrateTemplate(templateFilepath, outputConfigurationFilepath, databaseFilepath, serialNumberFilepath string) {
	fileCopy(templateFilepath, outputConfigurationFilepath)

	contentAsBytes, err := ioutil.ReadFile(outputConfigurationFilepath)
	if err != nil {
		fmt.Println("Error while reading " + outputConfigurationFilepath)
		fmt.Println(err)
	}

	//After copying the contents of the file needs to be altered to match input domain name
	contentsAsString := string(contentAsBytes[:])
	newFileContents := fmt.Sprintf(contentsAsString, databaseFilepath, serialNumberFilepath)

	ioutil.WriteFile(outputConfigurationFilepath, []byte(newFileContents), 0644)
}

//Usage: go run generate_certificates.go <domain.name>
//domain.name will be created as a directory and files generated by generate_certificates.go will go into the directory with name "domain.name".

//In the code, the term "server" refers to the computer hosting the name domain.name
func main() {
	//Force there to be exactly two arguments, the name of the file and the domain name
	if len(os.Args) != 2 {
		fmt.Println("Error: no domain name specified.")
		fmt.Println("usage: go run generate_certificates.go <domain.name>")
		os.Exit(0)
	}

	domainName := os.Args[1]
	domainNameDirectory := "output/" + domainName

	//Stage 1
	makeDirectories(domainNameDirectory)

	//Stage 2
	makePrivateKeys(domainNameDirectory)

	//Stage 3
	//1)Generate root certificate
	//2)Generate intermediate certificate
	//3)Generate server certificate
	// rootCertificate := fmt.Sprintf("")
	// if !fileExists(rootCertificate) {
	// 	fmt.Println("Generating root certificate from root private key.")
	// 	//generateCertificate()
	// 	// + rootAuthorityPrivateKey + " using configuration options from " +)
	// 	// openssl ca -selfsign -keyfile root.pem -config root_ca.conf -out root.crt -in root.csr -outdir root_certificates -verbose -batch
	// }
	// //root_csr.conf

	rootCSR := fmt.Sprintf("%s/root.csr", domainNameDirectory)
	if !fileExists(rootCSR) {
		fmt.Println("Generating root CSR.")

		rootAuthorityCSRConfigFilepath := domainNameDirectory + "/" + rootAuthorityCSRConfigFilename
		if !fileExists(rootAuthorityCSRConfigFilepath) {
			//Copy the config file over from the templates directory to the directoryNameDirectory if it doesn't exist
			templatesDirectoryRootAuthorityCSRConfigFilepath := templatesDirectory + "/" + rootAuthorityCSRConfigFilename
			fileCopy(templatesDirectoryRootAuthorityCSRConfigFilepath, rootAuthorityCSRConfigFilepath)
		}
		generateCertificateSigningRequest(domainNameDirectory+"/"+rootAuthorityPrivateKeyFilename, rootAuthorityCSRConfigFilepath, rootCSR)
	}

	rootCertificate := fmt.Sprintf("%s/root.crt", domainNameDirectory)
	if !fileExists(rootCertificate) {
		rootAuthorityConfigurationFilepath := rootAuthorityDirectory + "/" + rootAuthorityConfigurationFilename
		rootAuthorityDatabaseFilepath := rootAuthorityDirectory + "/" + rootAuthorityDatabaseFilename
		rootSerialNumberFilepath := rootAuthorityDirectory + "/" + rootAuthoritySerialNumberFilename

		if !fileExists(rootAuthorityConfigurationFilepath) {
			fmt.Println("No root authority configuration file detected in " + rootAuthorityConfigurationFilepath + ". Copying from templates directory")

			templatesDirectoryRootAuthorityConfigFilepath := templatesDirectory + "/" + rootAuthorityConfigurationFilename
			hydrateTemplate(templatesDirectoryRootAuthorityConfigFilepath, rootAuthorityConfigurationFilepath, rootAuthorityDatabaseFilepath, rootSerialNumberFilepath)
		}

		//Must also ensure the files referenced in the root authority configuration file exists

		if !fileExists(rootAuthorityDatabaseFilepath) {
			fmt.Println("Generating root database number file:" + rootAuthorityDatabaseFilepath)
			rootAuthorityDatabaseFile, error := os.Create(rootAuthorityDatabaseFilepath)
			if error != nil {
				fmt.Println("Error while creating root authority database file:" + rootAuthorityDatabaseFilepath)
				fmt.Println(error)
			}

			rootAuthorityDatabaseFile.Close()
		}

		if !fileExists(rootSerialNumberFilepath) {
			fmt.Println("Generating root serial number file:" + rootSerialNumberFilepath)
			rootSerialNumberFile, error := os.Create(rootSerialNumberFilepath)
			if error != nil {
				fmt.Println(error)
			}

			//Serial numbers file needs to have the hexadecimal digit 01 in it when initially created.
			rootSerialNumberFile.WriteString("01")
			rootSerialNumberFile.Close()
		}

		fmt.Println("Generating root certificate")
		generateSelfSignedCertificate(domainNameDirectory+"/"+rootAuthorityPrivateKeyFilename, rootCSR, rootAuthorityConfigurationFilepath, domainName, rootAuthorityDirectory, rootCertificate)
	}

	intermediateCSR := fmt.Sprintf("%s/intermediate.csr", domainNameDirectory)
	if !fileExists(intermediateCSR) {
		fmt.Println("Generating intermediate CSR.")

		intermediateAuthorityCSRConfigFilepath := domainNameDirectory + "/" + intermediateAuthorityCSRConfigFilename
		if !fileExists(intermediateAuthorityCSRConfigFilepath) {
			templatesDirectoryIntermediateAuthorityCSRConfigFilepath := templatesDirectory + "/" + intermediateAuthorityCSRConfigFilename
			fileCopy(templatesDirectoryIntermediateAuthorityCSRConfigFilepath, intermediateAuthorityCSRConfigFilename)
		}
		generateCertificateSigningRequest(domainNameDirectory+"/"+intermediateAuthorityPrivateKeyFilename, intermediateAuthorityCSRConfigFilepath, intermediateCSR)
	}

	intermediateCertificate := fmt.Sprintf("%s/intermediate.crt", domainNameDirectory)
	if !fileExists(intermediateCertificate) {
		intermediateAuthorityConfigurationFilepath := intermediateAuthorityDirectory + "/" + intermediateAuthorityConfigurationFilename
		intermediateAuthorityDatabaseFilepath := intermediateAuthorityDirectory + "/" + intermediateAuthorityDatabaseFilename
		intermediateAuthoritySerialNumberFilepath := intermediateAuthorityDirectory + "/" + intermediateAuthoritySerialNumberFilename

		if !fileExists(intermediateAuthorityConfigurationFilepath) {
			fmt.Println("No intermediate authority configuration file detected in " + intermediateAuthorityConfigurationFilepath + ". Copying from templates directory")

			templatesDirectoryIntermediateAuthorityConfigFilepath := templatesDirectory + "/" + intermediateAuthorityConfigurationFilename
			hydrateTemplate(templatesDirectoryIntermediateAuthorityConfigFilepath, intermediateAuthorityConfigurationFilepath, intermediateAuthorityDatabaseFilepath, intermediateAuthoritySerialNumberFilepath)
		}

		if !fileExists(intermediateAuthorityDatabaseFilepath) {
			fmt.Println("Generating root database number file:" + intermediateAuthorityDatabaseFilepath)
			intermediateAuthorityDatabaseFile, error := os.Create(intermediateAuthorityDatabaseFilepath)
			if error != nil {
				fmt.Println("Error while creating intermediate authority database file:" + intermediateAuthorityDatabaseFilepath)
				fmt.Println(error)
			}

			intermediateAuthorityDatabaseFile.Close()
		}

		if !fileExists(intermediateAuthoritySerialNumberFilepath) {
			fmt.Println("Generating intermediate authority serial number file:" + intermediateAuthoritySerialNumberFilepath)
			intermediateAuthoritySerialNumberFile, error := os.Create(intermediateAuthoritySerialNumberFilepath)
			if error != nil {
				fmt.Println(error)
			}

			//Serial numbers file needs to have the hexadecimal digit 01 in it when initially created.
			intermediateAuthoritySerialNumberFile.WriteString("01")
			intermediateAuthoritySerialNumberFile.Close()
		}

		fmt.Println("Generating intermediate certificate")
		// generateSignedCertificate(intermediateCSR, /**/outputCertificateFilename, certificateAuthorityConfiguration, certificateAuthoritySigningKey, certificateAuthorityCertificate, outputCertificateDirectory)
	}
}
