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
    // Use the running JVM directly; no toolchain auto-detection. Velocity API
    // 4.2.0 is compiled for Java 25, so the plugin targets Java 25 as well.
}

tasks.withType<JavaCompile>().configureEach {
    options.release = 25
}

tasks.jar {
    archiveFileName.set("minecraftserverlauncher-velocity-${project.version}.jar")
}
