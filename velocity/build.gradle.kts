plugins {
    `java-library`
}

group = "net.minecraftserverlauncher"
version = "0.1.0"

repositories {
    mavenCentral()
    // PaperMC hosts the Velocity API artifacts.
    maven("https://repo.papermc.io/repository/maven-public/")
}

dependencies {
    compileOnly("com.velocitypowered:velocity-api:4.2.0")
    annotationProcessor("com.velocitypowered:velocity-api:4.2.0")
}

java {
    // Use the running JVM (JDK 21 on this machine) directly; no toolchain
    // auto-detection. Bytecode targets Java 17 via the release flag below.
}

tasks.withType<JavaCompile>().configureEach {
    options.release = 25
}

tasks.jar {
    archiveFileName.set("minecraftserverlauncher-velocity-${project.version}.jar")
}
